# Sistem Pelaporan Prestasi Mahasiswa 

> **Proyek Ujian Akhir Semester (UAS)**
> Mata Kuliah: Pemrograman Backend Lanjut (Praktikum)
> D4 Teknik Informatika - Universitas Airlangga

---

| Atribut | Detail |
| :--- | :--- |
| **Nama** | **Putu Bagus Wedha Widagdha** |
| **NIM** | **434231044** |
| **Kelas** | **C6** |
| **Mata Kuliah** | Pemrograman Backend Lanjutan (Praktikum) |

---

## ✨ Fitur Utama

### 1. Autentikasi & Otorisasi (RBAC)
* Login, Refresh Token, dan Logout.
* Middleware untuk memvalidasi permission berdasarkan role (Admin, Mahasiswa, Dosen Wali)[cite: 169].

### 2. Manajemen Prestasi (Mahasiswa)
* **Input Dinamis:** Mendukung berbagai tipe prestasi seperti Akademik, Kompetisi, Organisasi, Publikasi, dan Sertifikasi[cite: 111].
* **Workflow:** Prestasi dimulai dari status `draft`, kemudian di-`submit` untuk verifikasi[cite: 96].
* **Upload Bukti:** Mendukung lampiran file bukti prestasi[cite: 147].

### 3. Verifikasi (Dosen Wali)
* Melihat daftar prestasi mahasiswa bimbingan.
* Melakukan **Approval** (Verified) atau **Rejection** (dengan catatan penolakan)[cite: 212, 222].

### 4. Manajemen User (Admin)
* CRUD User, assign Role, dan mapping data Mahasiswa ke Dosen Wali[cite: 235].

---

## 📂 Struktur Database

### PostgreSQL Schema
Menangani data inti dan relasi:
* `users`, `roles`, `permissions`, `role_permissions`
* `students`, `lecturers`
* `achievement_references` (Menyimpan status dan link ke MongoDB)[cite: 92].

### MongoDB Collection
Menangani detail prestasi (`achievements`):
* Menyimpan field dinamis seperti `rank`, `medalType` (untuk kompetisi), atau `publicationType`, `issn` (untuk publikasi) dalam satu dokumen JSON[cite: 114].

---

## 🔗 Dokumentasi API

Berikut adalah ringkasan endpoint utama yang tersedia:

| Method | Endpoint | Deskripsi | Akses |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/v1/auth/login` | Masuk ke sistem | Public |
| `GET` | `/api/v1/achievements` | List prestasi (Filter by role) | All |
| `POST` | `/api/v1/achievements` | Tambah prestasi baru | Mahasiswa |
| `POST` | `/api/v1/achievements/:id/submit` | Ajukan verifikasi | Mahasiswa |
| `POST` | `/api/v1/achievements/:id/verify` | Setujui prestasi | Dosen Wali |
| `POST` | `/api/v1/achievements/:id/reject` | Tolak prestasi | Dosen Wali |
| `GET` | `/api/v1/reports/statistics` | Statistik prestasi | All |

---

