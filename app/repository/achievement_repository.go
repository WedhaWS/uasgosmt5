package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"github.com/WedhaWS/uasgosmt5/app/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AchievementRepository struct {
	pgDB      *sql.DB
	mongoColl *mongo.Collection
}

func NewAchievementRepository(pg *sql.DB, mongoDB *mongo.Database) *AchievementRepository {
	return &AchievementRepository{
		pgDB:      pg,
		mongoColl: mongoDB.Collection("achievements"),
	}
}

// --- CREATE (HYBRID TRANSACTION) ---
func (r *AchievementRepository) Create(ctx context.Context, content *model.Achievement, ref *model.AchievementReference) error {
	now := time.Now()
	content.CreatedAt = now
	content.UpdatedAt = now
	ref.CreatedAt = now
	ref.UpdatedAt = now

	// 1. Insert ke MongoDB
	res, err := r.mongoColl.InsertOne(ctx, content)
	if err != nil {
		return err
	}

	// Ambil ID dari Mongo
	oid, _ := res.InsertedID.(primitive.ObjectID)
	ref.MongoAchievementID = oid.Hex()

	// 2. Insert ke PostgreSQL
	query := `
		INSERT INTO achievement_references (
			student_id, mongo_achievement_id, title, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	err = r.pgDB.QueryRow(
		query,
		ref.StudentID,
		ref.MongoAchievementID,
		ref.Title,
		ref.Status,
		ref.CreatedAt,
		ref.UpdatedAt,
	).Scan(&ref.ID)

	if err != nil {
		// KOMPENSASI (ROLLBACK MANUAL):
		// Jika simpan ke Postgres gagal, hapus data sampah di Mongo
		_, _ = r.mongoColl.DeleteOne(ctx, bson.M{"_id": oid})
		return errors.New("failed to save reference to postgres: " + err.Error())
	}

	return nil
}

// --- ADD ATTACHMENT ---
func (r *AchievementRepository) AddAttachment(ctx context.Context, refID string, attachment model.AchievementAttachment) error {
	// 1. Get MongoDB ID from PostgreSQL reference
	var mongoID string
	err := r.pgDB.QueryRow("SELECT mongo_achievement_id FROM achievement_references WHERE id = $1", refID).Scan(&mongoID)
	if err != nil {
		return errors.New("achievement reference not found")
	}

	// 2. Convert to ObjectID
	objID, err := primitive.ObjectIDFromHex(mongoID)
	if err != nil {
		return errors.New("invalid mongo id format")
	}

	// 3. Add attachment to MongoDB document
	_, err = r.mongoColl.UpdateOne(
		ctx,
		bson.M{"_id": objID},
		bson.M{"$push": bson.M{"attachments": attachment}},
	)

	return err
}

// --- FIND ALL (PAGINATION, SORT, SEARCH) ---
func (r *AchievementRepository) FindAll(param model.PaginationParam, studentID string, advisorID string) ([]model.AchievementReference, int64, error) {
	var achievements []model.AchievementReference
	var total int64

	// Base Query
	baseQuery := `
		SELECT 
			ar.id, ar.student_id, ar.mongo_achievement_id, ar.title, ar.status, 
			ar.submitted_at, ar.verified_at, ar.verified_by, ar.rejection_note, ar.created_at, ar.updated_at,
			u.full_name, s.student_id
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		JOIN users u ON s.user_id = u.id`

	var conditions []string
	var args []interface{}
	argId := 1

	// Filter by Student (RBAC)
	if studentID != "" {
		conditions = append(conditions, fmt.Sprintf("ar.student_id = $%d", argId))
		args = append(args, studentID)
		argId++
	}

	// Filter by Advisor (RBAC)
	if advisorID != "" {
		conditions = append(conditions, fmt.Sprintf("s.advisor_id = $%d", argId))
		args = append(args, advisorID)
		argId++
	}

	// Exclude soft deleted records
	conditions = append(conditions, fmt.Sprintf("ar.status != $%d", argId))
	args = append(args, "deleted")
	argId++

	// Search Logic (Title OR Status)
	if param.Search != "" {
		searchLike := "%" + strings.ToLower(param.Search) + "%"
		conditions = append(conditions, fmt.Sprintf("(LOWER(ar.title) LIKE $%d OR LOWER(ar.status) LIKE $%d)", argId, argId))
		args = append(args, searchLike)
		argId++ // Increment argId for the next potential argument
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	// 1. Count Total (Untuk Pagination)
	countQuery := `
		SELECT COUNT(*) 
		FROM achievement_references ar 
		JOIN students s ON ar.student_id = s.id 
	` + whereClause

	if err := r.pgDB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// 2. Sorting
	orderBy := "ar.created_at" // Default
	if param.SortBy != "" {
		validSorts := map[string]string{
			"title":      "ar.title",
			"status":     "ar.status",
			"created_at": "ar.created_at",
		}
		if val, ok := validSorts[param.SortBy]; ok {
			orderBy = val
		}
	}

	orderDir := "DESC"
	if strings.ToUpper(param.Order) == "ASC" {
		orderDir = "ASC"
	}

	// 3. Pagination
	limit := param.Limit
	offset := (param.Page - 1) * param.Limit

	// Final Query
	finalQuery := fmt.Sprintf(
		"%s %s ORDER BY %s %s LIMIT %d OFFSET %d",
		baseQuery, whereClause, orderBy, orderDir, limit, offset,
	)

	rows, err := r.pgDB.Query(finalQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var ar model.AchievementReference
		ar.Student = &model.Student{User: &model.User{}}

		var subAt, verAt sql.NullTime
		var verBy sql.NullString
		var rejNote sql.NullString

		err := rows.Scan(
			&ar.ID, &ar.StudentID, &ar.MongoAchievementID, &ar.Title, &ar.Status,
			&subAt, &verAt, &verBy, &rejNote, &ar.CreatedAt, &ar.UpdatedAt,
			&ar.Student.User.FullName, &ar.Student.StudentID,
		)
		if err != nil {
			return nil, 0, err
		}

		if subAt.Valid {
			ar.SubmittedAt = &subAt.Time
		}
		if verAt.Valid {
			ar.VerifiedAt = &verAt.Time
		}
		if verBy.Valid {
			str := verBy.String
			ar.VerifiedBy = &str
		}
		ar.RejectionNote = rejNote.String

		achievements = append(achievements, ar)
	}

	return achievements, total, nil
}

// --- FIND DETAIL (HYBRID FETCH) ---
func (r *AchievementRepository) FindDetail(ctx context.Context, id string) (*model.AchievementReference, *model.Achievement, error) {
	// 1. Ambil data Metadata dari Postgres
	query := `
		SELECT 
			ar.id, ar.student_id, ar.mongo_achievement_id, ar.title, ar.status, 
			ar.submitted_at, ar.verified_at, ar.verified_by, ar.rejection_note, ar.created_at, ar.updated_at,
			s.student_id, s.advisor_id, u.full_name,
			ver_u.full_name
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		JOIN users u ON s.user_id = u.id
		LEFT JOIN users ver_u ON ar.verified_by = ver_u.id
		WHERE ar.id = $1 AND ar.status != 'deleted'`

	var ref model.AchievementReference
	ref.Student = &model.Student{User: &model.User{}}
	ref.Verifier = &model.User{}

	var subAt, verAt sql.NullTime
	var verBy sql.NullString
	var rejNote sql.NullString
	var verifierName sql.NullString
	var advisorID sql.NullString

	err := r.pgDB.QueryRow(query, id).Scan(
		&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Title, &ref.Status,
		&subAt, &verAt, &verBy, &rejNote, &ref.CreatedAt, &ref.UpdatedAt,
		&ref.Student.StudentID, &advisorID, &ref.Student.User.FullName,
		&verifierName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, errors.New("achievement reference not found")
		}
		return nil, nil, err
	}

	if subAt.Valid {
		ref.SubmittedAt = &subAt.Time
	}
	if verAt.Valid {
		ref.VerifiedAt = &verAt.Time
	}
	if verBy.Valid {
		str := verBy.String
		ref.VerifiedBy = &str
		ref.Verifier.FullName = verifierName.String
	}
	if advisorID.Valid {
		str := advisorID.String
		ref.Student.AdvisorID = &str
	}
	ref.RejectionNote = rejNote.String

	// 2. Ambil data Detail dari MongoDB
	var content model.Achievement
	objID, err := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	if err != nil {
		return &ref, nil, errors.New("invalid mongo id format")
	}

	err = r.mongoColl.FindOne(ctx, bson.M{"_id": objID}).Decode(&content)
	if err != nil {
		return &ref, nil, errors.New("detail data not found in mongo")
	}

	return &ref, &content, nil
}

// --- UPDATE STATUS (VERIFIKASI DOSEN & UPDATE POIN) ---
func (r *AchievementRepository) UpdateStatus(id string, status string, verifiedBy string, note string, points int) error {
	// 1. Update di PostgreSQL (Status, VerifiedBy, RejectionNote)
	query := `
		UPDATE achievement_references 
		SET status = $1, updated_at = $2, verified_by = $3, verified_at = $4, rejection_note = $5
		WHERE id = $6
		RETURNING mongo_achievement_id`

	now := time.Now()

	var verBy interface{} = nil
	if verifiedBy != "" {
		verBy = verifiedBy
	}

	var mongoID string
	err := r.pgDB.QueryRow(query, status, now, verBy, now, note, id).Scan(&mongoID)
	if err != nil {
		return err
	}

	// 2. Update di MongoDB (Hanya jika Verified, kita simpan Poinnya)
	if status == "verified" && mongoID != "" {
		objID, err := primitive.ObjectIDFromHex(mongoID)
		if err == nil {
			// Update field "points" di dokumen MongoDB
			_, err = r.mongoColl.UpdateOne(
				context.Background(),
				bson.M{"_id": objID},
				bson.M{"$set": bson.M{"points": points}},
			)
			if err != nil {
				return errors.New("failed to update points in mongo: " + err.Error())
			}
		}
	}

	return nil
}

// --- SOFT DELETE ---
func (r *AchievementRepository) Delete(ctx context.Context, id string) error {
	var mongoID string
	var status string

	// 1. Get achievement info
	err := r.pgDB.QueryRow("SELECT mongo_achievement_id, status FROM achievement_references WHERE id = $1", id).Scan(&mongoID, &status)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("achievement not found")
		}
		return err
	}

	// 2. Check precondition: only draft can be deleted
	if status != "draft" {
		return errors.New("cannot delete submitted or verified achievement")
	}

	// 3. Soft delete in PostgreSQL (update status to 'deleted')
	now := time.Now()
	_, err = r.pgDB.Exec(
		"UPDATE achievement_references SET status = $1, updated_at = $2 WHERE id = $3",
		"deleted", now, id,
	)
	if err != nil {
		return errors.New("failed to soft delete reference: " + err.Error())
	}

	// 4. Soft delete in MongoDB (add deleted flag and timestamp)
	if mongoID != "" {
		objID, err := primitive.ObjectIDFromHex(mongoID)
		if err != nil {
			return errors.New("invalid mongo id format")
		}

		_, err = r.mongoColl.UpdateOne(
			ctx,
			bson.M{"_id": objID},
			bson.M{
				"$set": bson.M{
					"isDeleted": true,
					"deletedAt": now,
					"updatedAt": now,
				},
			},
		)
		if err != nil {
			return errors.New("failed to soft delete achievement: " + err.Error())
		}
	}

	return nil
}

// --- STATISTICS (FR-011) ---

type StatsResult struct {
	TotalPerType   map[string]int `json:"totalPerType"`
	TotalPerLevel  map[string]int `json:"totalPerLevel"`
	TotalPerPeriod map[string]int `json:"totalPerPeriod"`
	TopStudents    []TopStudent   `json:"topStudents"`
	Summary        StatsSummary   `json:"summary"`
}

type StatsSummary struct {
	TotalAchievements int `json:"totalAchievements"`
	TotalVerified     int `json:"totalVerified"`
	TotalPending      int `json:"totalPending"`
	TotalRejected     int `json:"totalRejected"`
	TotalPoints       int `json:"totalPoints"`
}

type TopStudent struct {
	Name        string `json:"name"`
	Program     string `json:"programStudy"`
	TotalPoints int    `json:"totalPoints"`
}

// GetStatistics generates overall stats
func (r *AchievementRepository) GetStatistics(ctx context.Context) (*StatsResult, error) {
	result := &StatsResult{
		TotalPerType:   make(map[string]int),
		TotalPerLevel:  make(map[string]int),
		TotalPerPeriod: make(map[string]int),
		TopStudents:    []TopStudent{},
		Summary:        StatsSummary{},
	}

	// 1. Total Per Type (Aggregation MongoDB)
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$achievementType"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}
	cursor, err := r.mongoColl.Aggregate(ctx, pipeline)
	if err == nil {
		var typeStats []struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err = cursor.All(ctx, &typeStats); err == nil {
			for _, s := range typeStats {
				result.TotalPerType[s.ID] = s.Count
			}
		}
	}

	// 2. Total Per Level (Aggregation MongoDB)
	pipelineLevel := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "achievementType", Value: "competition"}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$details.competitionLevel"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}
	cursorLevel, err := r.mongoColl.Aggregate(ctx, pipelineLevel)
	if err == nil {
		var levelStats []struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err = cursorLevel.All(ctx, &levelStats); err == nil {
			for _, s := range levelStats {
				key := s.ID
				if key == "" {
					key = "unknown"
				}
				result.TotalPerLevel[key] = s.Count
			}
		}
	}

	// 3. Top Students (Points from Mongo, Name from Postgres)
	// We aggregate studentID and their total points from Mongo first
	pipelineTop := mongo.Pipeline{
		// Only fetch achievements that have points (assuming > 0)
		{{Key: "$match", Value: bson.D{{Key: "points", Value: bson.D{{Key: "$gt", Value: 0}}}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$studentId"}, // Group by StudentID (UUID String)
			{Key: "totalPoints", Value: bson.D{{Key: "$sum", Value: "$points"}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "totalPoints", Value: -1}}}}, // Sort Descending
		{{Key: "$limit", Value: 5}},                                      // Top 5
	}

	cursorTop, err := r.mongoColl.Aggregate(ctx, pipelineTop)
	if err == nil {
		var topList []struct {
			StudentID   string `bson:"_id"`
			TotalPoints int    `bson:"totalPoints"`
		}

		if err = cursorTop.All(ctx, &topList); err == nil {
			// Loop aggregation results and get student name from Postgres
			for _, t := range topList {
				var name, program string

				// Query to students & users table
				// Ensure studentID is valid in Postgres
				queryUser := `
					SELECT u.full_name, s.program_study 
					FROM students s 
					JOIN users u ON s.user_id = u.id 
					WHERE s.id = $1`

				err := r.pgDB.QueryRow(queryUser, t.StudentID).Scan(&name, &program)
				if err != nil {
					// If user not found in Postgres (e.g. zombie data in Mongo), skip or set "Unknown"
					// fmt.Printf("Warning: Student ID %s not found in Postgres\n", t.StudentID)
					continue
				}

				result.TopStudents = append(result.TopStudents, TopStudent{
					Name:        name,
					Program:     program,
					TotalPoints: t.TotalPoints,
				})
			}
		}
	}

	// 4. Total Per Period (Last 6 months)
	pipelinePeriod := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "createdAt", Value: bson.D{
				{Key: "$gte", Value: time.Now().AddDate(0, -6, 0)}, // Last 6 months
			}},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "year", Value: bson.D{{Key: "$year", Value: "$createdAt"}}},
				{Key: "month", Value: bson.D{{Key: "$month", Value: "$createdAt"}}},
			}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}

	cursorPeriod, err := r.mongoColl.Aggregate(ctx, pipelinePeriod)
	if err == nil {
		var periodStats []struct {
			ID struct {
				Year  int `bson:"year"`
				Month int `bson:"month"`
			} `bson:"_id"`
			Count int `bson:"count"`
		}
		if err = cursorPeriod.All(ctx, &periodStats); err == nil {
			for _, s := range periodStats {
				key := fmt.Sprintf("%d-%02d", s.ID.Year, s.ID.Month)
				result.TotalPerPeriod[key] = s.Count
			}
		}
	}

	// 5. Summary Statistics (from PostgreSQL for accurate counts)
	var totalAchievements, totalVerified, totalPending, totalRejected int

	// Total achievements (exclude deleted)
	r.pgDB.QueryRow("SELECT COUNT(*) FROM achievement_references WHERE status != 'deleted'").Scan(&totalAchievements)

	// By status
	r.pgDB.QueryRow("SELECT COUNT(*) FROM achievement_references WHERE status = 'verified'").Scan(&totalVerified)
	r.pgDB.QueryRow("SELECT COUNT(*) FROM achievement_references WHERE status = 'submitted'").Scan(&totalPending)
	r.pgDB.QueryRow("SELECT COUNT(*) FROM achievement_references WHERE status = 'rejected'").Scan(&totalRejected)

	// Total points (from MongoDB)
	var totalPoints int
	pipelinePoints := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "points", Value: bson.D{{Key: "$gt", Value: 0}}}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "totalPoints", Value: bson.D{{Key: "$sum", Value: "$points"}}},
		}}},
	}

	cursorPoints, err := r.mongoColl.Aggregate(ctx, pipelinePoints)
	if err == nil {
		var pointsResult []struct {
			TotalPoints int `bson:"totalPoints"`
		}
		if err = cursorPoints.All(ctx, &pointsResult); err == nil && len(pointsResult) > 0 {
			totalPoints = pointsResult[0].TotalPoints
		}
	}

	result.Summary = StatsSummary{
		TotalAchievements: totalAchievements,
		TotalVerified:     totalVerified,
		TotalPending:      totalPending,
		TotalRejected:     totalRejected,
		TotalPoints:       totalPoints,
	}

	return result, nil
}

// GetStudentStatistics (Personal Stats)
func (r *AchievementRepository) GetStudentStatistics(ctx context.Context, studentID string) (*StatsResult, error) {
	result := &StatsResult{
		TotalPerType:   make(map[string]int),
		TotalPerLevel:  make(map[string]int),
		TotalPerPeriod: make(map[string]int),
		TopStudents:    []TopStudent{}, // Empty for individual stats
		Summary:        StatsSummary{},
	}

	// 1. Total Per Type
	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "studentId", Value: studentID}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$achievementType"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursor, err := r.mongoColl.Aggregate(ctx, pipeline)
	if err == nil {
		var stats []struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		cursor.All(ctx, &stats)
		for _, s := range stats {
			result.TotalPerType[s.ID] = s.Count
		}
	}

	// 2. Total Per Level (Competition only)
	pipelineLevel := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "studentId", Value: studentID},
			{Key: "achievementType", Value: "competition"},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$details.competitionLevel"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
	}

	cursorLevel, err := r.mongoColl.Aggregate(ctx, pipelineLevel)
	if err == nil {
		var levelStats []struct {
			ID    string `bson:"_id"`
			Count int    `bson:"count"`
		}
		if err = cursorLevel.All(ctx, &levelStats); err == nil {
			for _, s := range levelStats {
				key := s.ID
				if key == "" {
					key = "unknown"
				}
				result.TotalPerLevel[key] = s.Count
			}
		}
	}

	// 3. Total Per Period (Last 12 months for individual)
	pipelinePeriod := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "studentId", Value: studentID},
			{Key: "createdAt", Value: bson.D{
				{Key: "$gte", Value: time.Now().AddDate(-1, 0, 0)}, // Last 12 months
			}},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: bson.D{
				{Key: "year", Value: bson.D{{Key: "$year", Value: "$createdAt"}}},
				{Key: "month", Value: bson.D{{Key: "$month", Value: "$createdAt"}}},
			}},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "_id", Value: 1}}}},
	}

	cursorPeriod, err := r.mongoColl.Aggregate(ctx, pipelinePeriod)
	if err == nil {
		var periodStats []struct {
			ID struct {
				Year  int `bson:"year"`
				Month int `bson:"month"`
			} `bson:"_id"`
			Count int `bson:"count"`
		}
		if err = cursorPeriod.All(ctx, &periodStats); err == nil {
			for _, s := range periodStats {
				key := fmt.Sprintf("%d-%02d", s.ID.Year, s.ID.Month)
				result.TotalPerPeriod[key] = s.Count
			}
		}
	}

	// 4. Summary Statistics for this student
	var totalAchievements, totalVerified, totalPending, totalRejected int

	// Get counts from PostgreSQL
	r.pgDB.QueryRow("SELECT COUNT(*) FROM achievement_references WHERE student_id = $1 AND status != 'deleted'", studentID).Scan(&totalAchievements)
	r.pgDB.QueryRow("SELECT COUNT(*) FROM achievement_references WHERE student_id = $1 AND status = 'verified'", studentID).Scan(&totalVerified)
	r.pgDB.QueryRow("SELECT COUNT(*) FROM achievement_references WHERE student_id = $1 AND status = 'submitted'", studentID).Scan(&totalPending)
	r.pgDB.QueryRow("SELECT COUNT(*) FROM achievement_references WHERE student_id = $1 AND status = 'rejected'", studentID).Scan(&totalRejected)

	// Total points for this student
	var totalPoints int
	pipelinePoints := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "studentId", Value: studentID},
			{Key: "points", Value: bson.D{{Key: "$gt", Value: 0}}},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: nil},
			{Key: "totalPoints", Value: bson.D{{Key: "$sum", Value: "$points"}}},
		}}},
	}

	cursorPoints, err := r.mongoColl.Aggregate(ctx, pipelinePoints)
	if err == nil {
		var pointsResult []struct {
			TotalPoints int `bson:"totalPoints"`
		}
		if err = cursorPoints.All(ctx, &pointsResult); err == nil && len(pointsResult) > 0 {
			totalPoints = pointsResult[0].TotalPoints
		}
	}

	result.Summary = StatsSummary{
		TotalAchievements: totalAchievements,
		TotalVerified:     totalVerified,
		TotalPending:      totalPending,
		TotalRejected:     totalRejected,
		TotalPoints:       totalPoints,
	}

	return result, nil
}
