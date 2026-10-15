# SOFTWARE REQUIREMENTS SPECIFICATION (SRS)

## DevOpsGen: Developer-Oriented DevOps Scaffolding Tool

| Informasi | Detail |
|---|---|
| Nama proyek | DevOpsGen |
| Versi dokumen | 1.2.0 |
| Status | Approved for Implementation |
| Jenis aplikasi | Command Line Interface (CLI) dan Terminal User Interface (TUI) |
| Bahasa pemrograman | Go (Golang) |
| Lisensi | Apache License 2.0 |
| Target pengguna | Software Engineer, Backend Developer, DevOps Engineer, dan mahasiswa |
| Target platform | Linux, macOS Intel/Apple Silicon, dan Windows pada arsitektur yang dinyatakan didukung dan diuji |
| Tanggal baseline | 16 Oktober 2026 |

### Riwayat Revisi

| Versi | Tanggal | Perubahan |
|---|---|---|
| 1.0.0 | — | Baseline awal spesifikasi. |
| 1.1.0 | — | Penambahan matriks parameter, enum AppType, kontrak core engine, baseline benchmark, dan persyaratan validasi artefak. |
| 1.2.0 | 2026-10-16 | Finalisasi aturan resolusi konfigurasi, alur dan prioritas kegagalan, semantik test stage yang dilewati, semantik image/tag, keselamatan filesystem, rollback best-effort, serta perilaku port dan health probe. |

---

## 1. Pendahuluan

### 1.1 Tujuan

Dokumen ini mendefinisikan kebutuhan fungsional, nonfungsional, arsitektur, keamanan, pengujian, dan kriteria penerimaan DevOpsGen. Dokumen ini menjadi baseline bagi desain, implementasi, pengujian, dokumentasi, serta distribusi perangkat lunak.

DevOpsGen adalah *developer-oriented DevOps scaffolding tool* yang membantu pengembang menghasilkan konfigurasi containerization, CI/CD, dan deployment Kubernetes berdasarkan karakteristik proyek dan konfigurasi yang diberikan pengguna.

### 1.2 Latar Belakang

Pembuatan Dockerfile, workflow CI/CD, dan manifest Kubernetes secara manual dapat menimbulkan konfigurasi berulang, inkonsistensi, kesalahan sintaks, serta risiko operasional. DevOpsGen mengurangi pekerjaan tersebut dengan menghasilkan artefak yang terstandarisasi, dapat divalidasi, deterministik, dan dibuat dengan kebijakan filesystem yang aman.

Nilai utama DevOpsGen bukan sekadar membuat file konfigurasi. Tool ini harus mendukung proses yang dapat diuji, diotomatisasi, dan dikembangkan tanpa ketergantungan pada antarmuka terminal.

### 1.3 Ruang Lingkup

DevOpsGen menyediakan:

1. Mode interaktif berbasis TUI.
2. Mode non-interaktif untuk automation dan scripting.
3. Deteksi awal stack proyek.
4. Generasi Dockerfile.
5. Generasi GitHub Actions untuk pengujian dan publikasi container image.
6. Generasi manifest Kubernetes.
7. Validasi konfigurasi dan artefak.
8. Dry-run, no-overwrite, overwrite eksplisit, backup, serta best-effort rollback.
9. Distribusi executable binary untuk platform yang dinyatakan didukung dan diuji.

### 1.4 Tujuan Produk

DevOpsGen harus:

- Mengurangi kebutuhan menulis konfigurasi deployment dari nol.
- Menghasilkan artefak deterministik untuk input dan versi generator yang sama.
- Mencegah perubahan file tanpa persetujuan eksplisit.
- Memisahkan core engine dari CLI dan TUI.
- Mendukung automation dan pengujian tanpa credential eksternal.
- Memudahkan penambahan stack, template, dan provider di masa mendatang.
- Menyampaikan asumsi, nilai default, dan batasan hasil generasi secara transparan.

### 1.5 Definisi dan Singkatan

| Istilah | Definisi |
|---|---|
| CLI | Command Line Interface. |
| TUI | Terminal User Interface. |
| CI/CD | Continuous Integration / Continuous Delivery. |
| Stack | Bahasa pemrograman, framework, dan karakteristik aplikasi. |
| Registry | Layanan penyimpanan container image. |
| Artifact | File konfigurasi yang dihasilkan DevOpsGen. |
| Dry-run | Simulasi rencana perubahan tanpa menulis ke direktori target. |
| Headless mode | Mode non-interaktif yang tidak menunggu input terminal. |
| Deterministic output | Isi dan urutan artefak konsisten untuk versi serta input yang sama. |
| Rollback | Upaya mengembalikan perubahan yang dibuat oleh satu transaksi generasi. |
| Immutable tag | Tag yang mengidentifikasi build tertentu, dalam MVP berupa commit SHA. |
| Baseline | Nilai awal yang dapat disesuaikan, bukan klaim bahwa nilai tersebut optimal untuk semua aplikasi. |

---

## 2. Gambaran Umum Sistem

### 2.1 Perspektif Produk

DevOpsGen merupakan aplikasi mandiri yang dijalankan dari terminal di direktori proyek pengguna. Sistem membaca indikator proyek, menerima atau menyelesaikan konfigurasi, merender template di memori, memvalidasi artefak, menyusun rencana perubahan, lalu menulis hasil sesuai kebijakan filesystem.

Core engine tidak boleh bergantung pada Cobra, Bubble Tea, input terminal, layanan cloud, atau credential registry.

### 2.2 Target Pengguna

Target pengguna meliputi developer yang membutuhkan konfigurasi container dan CI/CD, DevOps engineer yang ingin menstandarkan scaffolding proyek, mahasiswa atau pemula yang mempelajari Docker dan Kubernetes, serta kontributor open-source yang membutuhkan generator konfigurasi yang dapat diotomatisasi.

Pengguna diasumsikan memahami dasar terminal dan struktur proyek perangkat lunak. Hasil generasi tetap perlu ditinjau sebelum digunakan pada lingkungan produksi.

### 2.3 Batasan MVP

MVP mencakup:

- Stack Go dan Node.js.
- GitHub Actions sebagai satu-satunya provider CI/CD.
- Docker Hub dan GitHub Container Registry (GHCR).
- Dockerfile.
- Manifest Kubernetes standar: Deployment dan Service, serta Ingress jika diminta.
- CLI, TUI, deteksi stack, validasi, dan filesystem safety.

Python, PHP/Laravel, Helm, Argo CD, Flux, provider CI/CD lain, serta mekanisme deployment otomatis berada di luar scope MVP.

### 2.4 Asumsi dan Dependensi

1. Pengguna memiliki akses baca ke direktori proyek.
2. Proses generasi membutuhkan izin tulis, kecuali dalam mode dry-run.
3. Deteksi stack bergantung pada indikator proyek yang tersedia dan tidak menjamin pemahaman penuh terhadap perilaku aplikasi.
4. Publikasi image membutuhkan konfigurasi autentikasi pada platform CI.
5. Validator eksternal bersifat opsional; validasi internal tetap wajib.
6. Ketersediaan Docker, Git, kubectl, dan tool Kubernetes tidak boleh diasumsikan tanpa pemeriksaan.
7. Pengujian core tidak boleh membutuhkan akses cloud atau credential sungguhan.
8. DevOpsGen tidak menjalankan build container, publikasi image, atau deployment sebagai efek samping dari generasi.

---

## 3. Dukungan Stack dan Deteksi

### 3.1 Stack Go

Sistem mendeteksi kandidat proyek Go melalui keberadaan `go.mod`. Keberadaan file tersebut merupakan indikator awal, bukan bukti bahwa direktori pasti dapat dibangun.

Dockerfile yang dihasilkan harus:

- Menggunakan multi-stage build.
- Menjalankan kompilasi pada tahap build.
- Memisahkan tool build dari runtime jika sesuai.
- Menggunakan runtime minimal yang kompatibel dengan hasil build.
- Mengekspos port konfigurasi sebagai metadata container.
- Menggunakan non-root user apabila kompatibel dengan aplikasi dan runtime.
- Memperhitungkan dependensi runtime, sertifikat CA, dan kebutuhan file aplikasi.

Pilihan runtime dapat menggunakan distroless atau Alpine jika sesuai. Template tidak boleh mengasumsikan binary statis atau kompatibilitas libc tanpa dasar yang jelas.

Nilai port atau instruksi `EXPOSE` tidak membuat aplikasi otomatis mendengarkan pada port tersebut. Aplikasi wajib mengikat listener ke port yang benar-benar dikonfigurasikan; DevOpsGen tidak mengubah source code aplikasi.

### 3.2 Stack Node.js

Sistem mendeteksi kandidat proyek Node.js melalui keberadaan `package.json` dan memeriksa:

- `dependencies` dan `devDependencies`.
- `scripts` yang tersedia.
- Lockfile seperti `package-lock.json`, `pnpm-lock.yaml`, atau `yarn.lock`.
- Indikator framework dan konfigurasi yang dapat dibaca secara statis.

Framework yang dapat dikenali mencakup Express, Next.js, Vite, dan NestJS. Deteksi hanya menghasilkan kandidat; indikator yang saling bertentangan harus dianggap ambigu.

#### 3.2.1 Resolusi Perintah

Resolusi build dan startup mengikuti urutan berikut:

1. Nilai yang diberikan secara eksplisit oleh pengguna.
2. Script proyek yang terdeteksi dan memenuhi kebutuhan stack. Dalam TUI, nilai ditampilkan untuk dikonfirmasi; dalam mode non-interaktif, kandidat unik yang valid dapat digunakan, kecuali pengguna menggantinya melalui flag.
3. Default khusus framework hanya jika didukung oleh indikator yang memadai dan sesuai dengan template yang dipilih.
4. Jika tidak ada pilihan yang dapat dipastikan, minta nilai eksplisit atau hentikan proses dengan pesan kesalahan yang jelas.

Keberadaan script tidak menjamin script itu benar untuk produksi. DevOpsGen tidak boleh mengklaim telah membuktikan perilaku aplikasi hanya dari isi `package.json`.

#### 3.2.2 Heuristik Perintah

| Indikator | Perilaku |
|---|---|
| `scripts.build` tersedia | Menjadi kandidat build command. |
| `scripts.test` tersedia | Menjadi kandidat test command. |
| `scripts.start` tersedia | Menjadi kandidat startup command. |
| Framework dapat diidentifikasi dengan meyakinkan | Sarankan template framework yang sesuai. |
| Framework atau runtime ambigu | Minta klarifikasi atau hentikan mode non-interaktif. |
| Port tidak memiliki bukti yang cukup | Minta port eksplisit. |
| Perintah wajib tidak tersedia | Minta konfigurasi eksplisit atau hentikan proses. |
| Lockfile dikenali | Pilih prosedur instalasi dependency yang sesuai. |
| Beberapa lockfile bertentangan | Laporkan ambiguitas; jangan memilih package manager secara diam-diam. |

### 3.3 Jenis Aplikasi (AppType)

| Stack | Nilai | Perilaku template |
|---|---|---|
| Go | `binary` | Menghasilkan satu binary aplikasi dan menjalankan binary tersebut secara langsung. |
| Node.js | `express` | Aplikasi server HTTP Express. Startup harus sesuai script/perintah yang tervalidasi. |
| Node.js | `nextjs` | Build Next.js dan runtime `next start` atau mode standalone jika konfigurasi proyek mendukungnya. |
| Node.js | `vite` | Build aset statis dan sajikan menggunakan template static server. Startup Node.js tidak diperlukan jika template static server yang dipilih tidak menggunakannya. |
| Node.js | `nestjs` | Build NestJS dan jalankan output yang benar-benar dihasilkan proyek, misalnya file di `dist`, setelah dikonfirmasi. |
| Node.js | `generic` | Aplikasi Node.js tanpa framework teridentifikasi; perintah build dan startup harus diberikan secara eksplisit jika dibutuhkan. |

Nilai di luar daftar ditolak dengan exit code `2`. Jika deteksi memberikan lebih dari satu kandidat yang masuk akal, mode interaktif meminta pilihan; mode non-interaktif mengembalikan exit code `5`.

### 3.4 Deteksi Ambigu

Jika beberapa stack terdeteksi atau indikator bertentangan, sistem harus melaporkan file dan indikator yang menyebabkan ambiguitas. Mode interaktif meminta pilihan pengguna. Mode non-interaktif berhenti dengan exit code `5` dan menyebut flag yang diperlukan untuk menyelesaikan ambiguitas.

Deteksi bersifat read-only dan tidak boleh mengubah file proyek.

---

## 4. Kebutuhan Fungsional

### 4.1 Matriks Parameter Konfigurasi

Parameter berikut menjadi kontrak konfigurasi MVP. “Wajib” berarti proses tidak dapat dilanjutkan tanpa nilai yang valid. Nilai dapat berasal dari flag eksplisit atau deteksi yang memenuhi aturan pada bagian 3. Nilai yang tidak dapat dipastikan harus diminta secara eksplisit.

| Parameter | Go | Node.js | Keterangan |
|---|---|---|---|
| `stack` | Wajib | Wajib | Enum `go` atau `nodejs`; dapat berasal dari deteksi jika hasilnya tunggal dan jelas. |
| `ci` | Wajib | Wajib | MVP: `github-actions`. |
| `registry` | Wajib | Wajib | `dockerhub` atau `ghcr`. |
| `image` | Wajib | Wajib | Repository image dalam bentuk `namespace/repository`, tanpa hostname registry dan tanpa tag. Contoh: `bangkah/my-app`. |
| `port` | Wajib | Wajib | Rentang 1–65535. Harus sesuai dengan port listener aplikasi. Jangan menebak dari port umum framework. |
| `build-cmd` | Opsional | Opsional bersyarat | Go: default template yang teruji. Node.js: gunakan kandidat yang terdeteksi atau default framework yang tervalidasi; bila tidak tersedia, wajib eksplisit. |
| `test-cmd` | Opsional | Opsional | Jika tidak tersedia, test stage dilewati secara transparan dan bukan dilaporkan sebagai berhasil. |
| `start-cmd` | Tidak berlaku | Wajib bersyarat | Go menjalankan binary. Node.js memerlukan perintah untuk server runtime; template Vite static server dapat memakai entrypoint server bawaan tanpa `start-cmd` Node.js. |
| `app-type` | Opsional | Opsional | Enum pada bagian 3.3; nilai default hanya jika deteksi jelas. |
| `tag` | Opsional | Opsional | Tag alias tambahan; default alias adalah `latest`. Workflow wajib menerbitkan tag commit SHA dan alias ini. Tag `latest` tidak boleh menjadi satu-satunya tag. |
| `k8s` | Opsional | Opsional | Default: menghasilkan Deployment dan Service. Dapat dinonaktifkan secara eksplisit. |
| `ingress` | Opsional | Opsional | Default nonaktif. Jika aktif, wajib diikuti `ingress-host`. |
| `ingress-host` | Kondisional | Kondisional | Wajib jika Ingress diaktifkan; harus memenuhi validasi hostname RFC 1123. |
| `health-path` | Opsional bersyarat | Opsional bersyarat | Endpoint HTTP pemeriksaan kesehatan yang diberikan pengguna. Tidak boleh ditebak. Jika tidak diberikan, HTTP readiness/liveness probe tidak dibuat. |
| `output-dir` | Opsional | Opsional | Default: direktori kerja saat ini. Path harus divalidasi. |
| `force` | Flag | Flag | Persetujuan eksplisit untuk mengganti file reguler yang konflik. Tidak mengizinkan penggantian symlink atau direktori. |
| `dry-run` | Flag | Flag | Simulasi tanpa menulis atau membuat file/direktori sementara pada target. |

`image` harus menyertakan namespace agar referensi image untuk registry dan Kubernetes tidak bergantung pada tebakan. Referensi image lengkap dibentuk sebagai `docker.io/<namespace>/<repository>` untuk Docker Hub atau `ghcr.io/<namespace>/<repository>` untuk GHCR. Nama dan komponen path harus sesuai aturan registry.

Pada mode non-interaktif, jika parameter wajib yang tidak dapat dideteksi belum disediakan, proses harus gagal dengan exit code `2`. Stack yang ambigu atau tidak didukung menggunakan exit code `5` sesuai aturan pada FR-16.

### 4.2 Resolusi Nilai dan Default

Sistem harus menggunakan urutan prioritas yang konsisten:

1. Nilai valid yang diberikan secara eksplisit melalui flags.
2. Nilai hasil deteksi yang tunggal, relevan, dan didukung indikator proyek.
3. Default template yang terdokumentasi dan sesuai dengan `stack` serta `app-type`.
4. Jika tetap tidak dapat ditentukan, hentikan proses dan minta nilai eksplisit.

Tidak ada default yang boleh menutupi ambiguitas stack, menciptakan port aplikasi, atau mengarang endpoint kesehatan. Dalam TUI, kandidat hasil deteksi ditampilkan dan dapat diganti sebelum generasi.

### FR-01: Inisialisasi

Sistem menyediakan:

`devopsgen init`

Perintah memulai TUI apabila terminal interaktif tersedia dan konfigurasi belum lengkap. Jika konfigurasi lengkap diberikan melalui flags, sistem menjalankan mode non-interaktif.

Jika TUI dibutuhkan tetapi stdin atau terminal tidak interaktif, proses tidak boleh menunggu input. Sistem mengembalikan kesalahan yang dapat ditindaklanjuti beserta parameter yang belum tersedia.

### FR-02: Mode Interaktif

Wizard harus mendukung seluruh parameter pada matriks 4.1, termasuk stack dan `app-type`, perintah build/test/start yang relevan, port, CI provider, registry, image name, tag, opsi Kubernetes, Ingress host, health path jika digunakan, serta direktori tujuan.

Input harus divalidasi sebelum diteruskan ke generator. Wizard menampilkan kandidat hasil deteksi sebagai kandidat dan meminta konfirmasi jika perilakunya tidak dapat dibuktikan secara statis. Pengguna dapat meninjau ringkasan konfigurasi sebelum menyetujui generasi.

### FR-03: Mode Non-interaktif

Contoh:

```bash
devopsgen init --stack go --ci github-actions --registry ghcr --image bangkah/my-app --port 8080
```

Sistem harus:

- Memproses konfigurasi melalui flags.
- Memastikan parameter wajib tersedia sesuai matriks 4.1.
- Tidak menunggu input stdin.
- Menolak konfigurasi yang tidak valid atau tidak lengkap.
- Menghasilkan exit code non-zero saat gagal.
- Menyebut parameter yang hilang atau tidak valid dalam pesan kesalahan.
- Tidak menjalankan perintah build, push, atau deploy sebagai bagian dari generasi.

Nilai command yang dimasukkan pengguna dapat ditulis ke Dockerfile atau workflow yang dihasilkan. DevOpsGen tidak mengeksekusi command tersebut secara lokal. Namun, command itu akan dieksekusi ketika pengguna membangun image atau menjalankan workflow; pengguna bertanggung jawab meninjau command sebelum penggunaan.

### FR-04: Deteksi Stack

Perintah:

`devopsgen detect`

Sistem harus melaporkan stack kandidat, indikator yang ditemukan, kandidat `app-type`, build/test/start command, lockfile/package manager, serta ambiguitas yang belum terselesaikan. Jika port tidak dapat ditentukan dengan bukti eksplisit, sistem harus menyatakannya dan tidak mengisi port berdasarkan asumsi framework.

Deteksi tidak boleh mengubah filesystem atau menjalankan script proyek.

### FR-05: Generasi Dockerfile

Sistem harus menghasilkan Dockerfile sesuai stack, `app-type`, dan konfigurasi.

Dockerfile harus:

- Menggunakan strategi build yang sesuai.
- Menghindari secret hardcoded.
- Meminimalkan komponen runtime yang tidak diperlukan.
- Menjalankan aplikasi sebagai non-root jika memungkinkan.
- Mengekspos port konfigurasi sebagai metadata, tanpa mengklaim bahwa `EXPOSE` mengubah port listener aplikasi.
- Menggunakan command build/start yang diselesaikan melalui aturan 4.2.
- Menghasilkan konten deterministik.

Pengecualian dari non-root hanya diperbolehkan jika terdapat kebutuhan kompatibilitas yang jelas dan alasan tersebut didokumentasikan. Template wajib memperhitungkan permission file, direktori kerja, CA certificates, dan kebutuhan runtime yang diketahui.

### FR-06: Generasi GitHub Actions

Workflow harus mencakup langkah yang relevan untuk:

1. Checkout repository.
2. Setup runtime yang diperlukan oleh stack.
3. Instalasi dependency untuk langkah test yang berjalan di runner.
4. Menjalankan test command jika tersedia atau dikonfigurasi.
5. Build container image.
6. Login ke registry setelah prasyarat terpenuhi.
7. Publish image setelah test dan build yang diwajibkan berhasil.

Workflow harus selalu menerbitkan tag commit SHA `${{ github.sha }}` sebagai tag immutable yang dapat ditelusuri ke commit sumber. Workflow juga menerbitkan alias `latest` secara default, atau alias kustom dari parameter `tag` jika diberikan. Tag commit SHA tetap wajib diterbitkan, sehingga alias mutable tidak pernah menjadi satu-satunya tag.

Jika test command tidak tersedia, workflow harus menyatakan dengan jelas bahwa test stage dilewati karena belum dikonfigurasi. Workflow harus menambahkan ringkasan atau notice yang eksplisit, tidak boleh membuat langkah test palsu, dan tidak boleh melaporkan bahwa pengujian berhasil jika pengujian tidak dijalankan. Status workflow yang sukses hanya berarti langkah yang benar-benar dikonfigurasi telah berhasil.

Workflow harus memakai permission minimum yang diperlukan, menghindari token yang tidak digunakan, dan tidak menjalankan publish sebelum prasyarat test/build yang dikonfigurasi berhasil. Penggunaan action pihak ketiga harus memakai referensi versi yang eksplisit dan tidak bergantung pada branch bergerak seperti `main` atau `master`.

### FR-07: Registry dan Autentikasi

#### Docker Hub

Workflow menggunakan GitHub Actions Secrets berikut:

- `DOCKER_USERNAME`
- `DOCKER_PASSWORD`

Image yang dihasilkan mengikuti bentuk `docker.io/<namespace>/<repository>`. Jika namespace image tidak sesuai dengan akun yang diizinkan untuk publish, workflow akan gagal dan pesan harus menjelaskan konfigurasi yang perlu diperiksa tanpa mencetak credential.

#### GitHub Container Registry (GHCR)

Workflow menggunakan `GITHUB_TOKEN` dengan permission minimum:

- `contents: read`
- `packages: write`

Image mengikuti bentuk `ghcr.io/<namespace>/<repository>`. Konfigurasi namespace harus kompatibel dengan pemilik repository dan permission workflow.

Credential tidak boleh ditulis sebagai plaintext ke artefak, argumen log, atau pesan kesalahan. Pengujian generasi tidak memerlukan credential sungguhan.

### FR-08: Generasi Manifest Kubernetes

Sistem menghasilkan:

- `deployment.yaml`
- `service.yaml`
- `ingress.yaml` jika diminta.

Tidak ada perintah `kubectl apply`, publikasi image, atau perubahan cluster yang boleh dijalankan selama generasi.

#### 4.8.1 Deployment

Validator minimum harus memastikan:

- `apiVersion` yang sesuai, `kind: Deployment`, dan `metadata.name`.
- `spec.selector.matchLabels` tersedia.
- `spec.template.metadata.labels` tersedia.
- `spec.template.spec.containers` berisi setidaknya satu container.
- Nama container valid dan unik dalam Pod.
- Image container tidak kosong.
- Port container sesuai konfigurasi apabila ditentukan.
- Selector cocok dengan label pada template Pod.

#### 4.8.2 Service

Validator minimum harus memastikan:

- `apiVersion` yang sesuai, `kind: Service`, dan `metadata.name`.
- `spec.selector` tersedia.
- `spec.ports` berisi port dan target port yang valid.
- Selector cocok dengan label Pod Deployment.
- `targetPort` mengarah ke port aplikasi/container yang benar-benar dikonfigurasi.

#### 4.8.3 Ingress

Jika diaktifkan, Ingress menggunakan API `networking.k8s.io/v1` dan harus mempunyai:

- `kind: Ingress`, `metadata.name`, dan aturan routing yang valid.
- Host yang lolos validasi hostname RFC 1123.
- Path dan `pathType` yang valid.
- Backend Service yang terdefinisi.
- Port backend yang cocok dengan port Service terkait.

Dokumentasi hasil generasi harus menyatakan bahwa Ingress membutuhkan Ingress Controller dan konfigurasi cluster yang sesuai. Keberadaan manifest Ingress saja tidak membuat akses eksternal otomatis tersedia.

#### 4.8.4 Resource dan Health Probe

Manifest awal harus menyediakan requests dan limits resource sebagai baseline yang dapat disesuaikan. Nilai default MVP adalah `requests.cpu: 100m`, `requests.memory: 128Mi`, `limits.cpu: 500m`, dan `limits.memory: 256Mi`. Nilai ini hanya baseline awal, bukan nilai optimal universal; pengguna harus mengevaluasi dan menyesuaikannya berdasarkan beban aplikasi serta batas cluster.

HTTP readiness/liveness probe hanya dibuat jika pengguna memberikan `health-path` atau detektor menemukan konfigurasi endpoint yang eksplisit dan dapat dipercaya. DevOpsGen tidak boleh mengarang endpoint seperti `/health` atau `/ready`. Jika endpoint pemeriksaan tidak diketahui, probe HTTP dihilangkan dan ringkasan generasi menjelaskan alasannya.

Container sebaiknya berjalan sebagai non-root jika kompatibel. Port pada manifest harus konsisten dengan konfigurasi pengguna. Pengguna tetap bertanggung jawab memastikan aplikasi benar-benar mendengarkan di port itu.

#### 4.8.5 Image pada Manifest

Deployment menggunakan nama image yang berasal dari registry dan `image` pada konfigurasi. Tag manifest harus sama dengan alias yang diterbitkan workflow: `latest` secara default atau nilai parameter `tag` jika diberikan. Workflow juga harus menerbitkan tag commit SHA. Dokumentasi harus menyarankan penggunaan tag immutable atau digest untuk deployment produksi. Dokumentasi harus menyarankan penggunaan tag immutable atau digest untuk deployment produksi.

### FR-09: Validasi Input

Sistem harus memvalidasi:

- Field wajib dan tipe data.
- Nama repository image dan namespace sesuai aturan registry.
- Port dalam rentang 1–65535.
- CI provider dan registry terhadap nilai yang didukung.
- Path output, nama file artefak, dan path traversal.
- Nilai `app-type` terhadap enum bagian 3.3.
- Konsistensi stack, runtime, build/start command, port, dan opsi Kubernetes.
- Hostname Ingress berdasarkan aturan RFC 1123.
- Perintah kustom sebagai string satu baris tanpa karakter kontrol yang dapat merusak struktur artefak.

Command kustom tetap merupakan kode yang dipercaya pengguna dan dapat dieksekusi ketika Dockerfile atau workflow dijalankan. DevOpsGen harus menghindari penyisipan tidak sengaja ke struktur YAML/Dockerfile dan mendokumentasikan risiko command kustom.

Nama resource Kubernetes mengikuti aturan nama resource terkait; aturan nama image mengikuti aturan registry, bukan aturan hostname atau resource Kubernetes.

### FR-10: Dry-run

Perintah:

`devopsgen init --dry-run`

Dry-run harus:

- Merender dan memvalidasi semua artefak dengan jalur kode yang sama seperti generasi normal.
- Menampilkan artefak yang akan dibuat atau diganti.
- Melaporkan konflik dan aksi yang direncanakan.
- Tidak menulis atau mengubah file target.
- Tidak membuat file atau direktori sementara pada direktori target.
- Tidak mengubah permission, timestamp, atau metadata file proyek.

Jika terdapat konflik tanpa `--force`, dry-run mengembalikan exit code `3` setelah melaporkan konflik. Jika `--force` juga diberikan, dry-run hanya melaporkan rencana penggantian dan tetap tidak menulis.

### FR-11: No-overwrite dan Force

Secara default, sistem tidak boleh menimpa file yang sudah ada. Dalam mode interaktif, sistem dapat meminta konfirmasi. Dalam mode non-interaktif, konflik tanpa `--force` menghasilkan exit code `3`.

Flag `--force` merupakan persetujuan eksplisit untuk mengganti file reguler yang ditargetkan. Flag ini tidak mengizinkan penggantian symlink, direktori, atau path yang keluar dari direktori target. Sebelum commit, sistem harus menyiapkan backup seluruh file lama yang akan diganti. Jika backup tidak dapat disiapkan dengan aman, transaksi dibatalkan sebelum file target dimodifikasi.

### FR-12: Transaksi Filesystem dan Best-effort Rollback

Sebelum commit, sistem harus:

1. Memvalidasi konfigurasi.
2. Merender seluruh artefak di memori.
3. Memvalidasi semua artefak.
4. Memvalidasi path dan memastikan seluruh tujuan berada dalam direktori target.
5. Mendeteksi konflik.
6. Menyiapkan rencana perubahan dan backup untuk file yang akan diganti.
7. Menyiapkan staging jika dibutuhkan, setelah dry-run dipastikan tidak aktif.

Commit harus mencatat setiap file baru, file lama yang diganti, backup, dan direktori yang dibuat oleh transaksi. File hasil rendering harus disiapkan sebelum dipindahkan ke tujuan. Sistem sebaiknya memakai rename pada filesystem yang sama jika tersedia, tetapi tidak boleh mengklaim atomicity absolut untuk transaksi multi-file lintas platform.

Jika terjadi kegagalan I/O, sistem harus berupaya:

- Menghapus file baru yang dibuat transaksi.
- Memulihkan file lama yang diganti menggunakan backup.
- Menghapus direktori baru yang kosong dan dibuat oleh transaksi.
- Mempertahankan file lama yang tidak dikelola transaksi.
- Melaporkan dengan jelas setiap langkah rollback yang gagal dan file yang memerlukan pemulihan manual.

Lokasi dan kebijakan pembersihan backup harus ditentukan oleh implementasi dan didokumentasikan. Rollback hanya boleh memengaruhi perubahan yang dapat dikaitkan dengan transaksi saat ini. Jika rollback gagal, exit code akhir adalah `1` dan pesan kesalahan harus membedakan kegagalan awal dari kegagalan pemulihan.

Filesystem tidak menyediakan transaksi multi-file portabel yang universal. Jaminan produk adalah best-effort rollback, bukan atomicity absolut.

### FR-13: Validasi Artefak

Perintah:

`devopsgen validate [--dir PATH]`

Secara default, sistem memvalidasi artefak yang didukung pada direktori kerja. Sistem harus memeriksa:

- Keberadaan file yang diwajibkan untuk artefak yang terdeteksi.
- Sintaks YAML.
- Field wajib resource Kubernetes.
- Konsistensi selector dan label.
- Image container.
- Relasi Service dan Ingress.
- Struktur dasar GitHub Actions workflow.
- Konsistensi port dan health probe jika dikonfigurasi.

Validasi internal wajib berjalan walaupun validator eksternal tidak tersedia. Validator eksternal bersifat tambahan dan tidak menggantikan validasi internal. Hasil harus membedakan status valid, tidak valid, dan pemeriksaan eksternal yang dilewati.

Contoh pesan:

`Internal validation passed; external validation skipped: kubeconform not found.`

Jika validasi artefak gagal, exit code adalah `4`. Jika alat eksternal tidak tersedia tetapi validasi internal berhasil, kondisi itu sendiri bukan kegagalan validasi.

### FR-14: Daftar Dukungan

Perintah:

`devopsgen list`

Sistem menampilkan stack, app type, registry, provider CI, dan template deployment yang tersedia dalam versi binary yang sedang berjalan. Daftar tersebut harus berasal dari kapabilitas aktual yang dibundel, bukan daftar roadmap masa depan.

### FR-15: Informasi Versi

Perintah:

`devopsgen version`

Sistem menampilkan versi aplikasi dan versi template. Pada MVP, versi template mengikuti versi binary. Informasi versi ditentukan saat build atau release; build development boleh menampilkan nilai seperti `dev` jika versi release tidak tersedia.

### FR-16: Help dan Exit Code

Setiap command harus menyediakan bantuan penggunaan yang konsisten. Kode keluar yang didukung:

| Exit code | Arti |
|---:|---|
| `0` | Operasi berhasil. |
| `1` | Kesalahan umum atau I/O, termasuk kegagalan rollback. |
| `2` | Input, flag, atau konfigurasi tidak valid atau tidak lengkap. |
| `3` | Konflik file tanpa persetujuan overwrite. |
| `4` | Validasi artefak gagal. |
| `5` | Stack ambigu atau tidak didukung. |

#### 4.16.1 Urutan Tahap dan Penentuan Exit Code

Implementasi memakai alur fail-fast yang deterministik. Kesalahan pada tahap yang lebih awal menghentikan proses; tahap berikutnya tidak dijalankan. Urutan tahap adalah:

1. Parsing flags dan validasi nilai eksplisit. Kesalahan menghasilkan `2`.
2. Deteksi stack jika diperlukan. Stack ambigu atau tidak didukung menghasilkan `5`.
3. Penyelesaian serta validasi konfigurasi lengkap. Parameter wajib yang masih hilang atau tidak konsisten menghasilkan `2`.
4. Rendering artefak dan validasi output. Kegagalan validasi artefak menghasilkan `4`.
5. Pemeriksaan path fisik, rencana filesystem, dan konflik. Path yang tidak aman menghasilkan `2`; konflik tanpa persetujuan menghasilkan `3`.
6. Backup, commit, dan operasi filesystem. Kesalahan I/O atau rollback menghasilkan `1`.

Aturan ini adalah *precedence by processing stage*, bukan klaim bahwa aplikasi harus menemukan seluruh kemungkinan kegagalan sebelum berhenti. Sebagai contoh, input invalid menghasilkan `2` tanpa melanjutkan ke pemeriksaan konflik; validasi artefak yang gagal menghasilkan `4` sebelum konflik diperiksa. Pengujian harus membuktikan perilaku short-circuit tersebut secara eksplisit.

---

## 5. Kebutuhan Nonfungsional

### NFR-01: Performa

Target rendering dan validasi internal adalah kurang dari satu detik pada benchmark referensi dengan ketentuan:

- Maksimal 20 artefak.
- Total ukuran output maksimal 256 KiB.
- Tidak menjalankan proses eksternal.
- Tidak mengakses jaringan.

Benchmark harus mencatat spesifikasi mesin, versi Go, perintah benchmark, dan kondisi pengujian. Waktu interaksi pengguna, operasi filesystem yang lambat, build image, publish, serta validasi eksternal tidak termasuk target core engine.

### NFR-02: Portabilitas

DevOpsGen harus dapat dibangun untuk target Linux, macOS Intel, macOS Apple Silicon, dan Windows pada arsitektur yang ditetapkan dalam pipeline release. Binary release tidak boleh membutuhkan instalasi Go pada mesin pengguna.

Kompatibilitas dibuktikan dengan build pada target yang ditentukan dan pengujian pada runner native jika tersedia. Cross-compilation saja tidak membuktikan bahwa binary berfungsi benar pada semua platform.

### NFR-03: Maintainability

CLI, TUI, deteksi, konfigurasi, generator, template, validator, dan filesystem harus dipisahkan dalam modul dengan tanggung jawab jelas. Penambahan stack baru tidak boleh mengharuskan perubahan pada kebijakan dasar filesystem atau kode TUI selain integrasi yang diperlukan.

### NFR-04: Determinisme dan Template Versioning

Versi template mengikuti versi binary pada MVP. Template dibundel menggunakan `embed.FS`, sehingga setiap binary release membawa kumpulan template yang tetap.

Untuk versi binary dan konfigurasi yang sama:

- Isi setiap artefak harus identik.
- Urutan artefak yang dilaporkan harus konsisten.
- Output tidak boleh bergantung pada urutan iterasi map yang tidak deterministik.
- Nilai lingkungan runtime tidak boleh disisipkan ke file hasil generasi kecuali menjadi bagian eksplisit dari konfigurasi.

Perubahan template yang memengaruhi output harus disertai pengujian dan catatan changelog.

### NFR-05: Keamanan

Sistem harus:

- Menghindari secret plaintext.
- Menggunakan permission CI yang minimum.
- Mencegah path traversal dan menolak jalur symlink yang keluar dari target.
- Menolak overwrite tanpa persetujuan.
- Membatasi penulisan pada direktori target yang divalidasi.
- Tidak menjalankan deployment otomatis.
- Tidak membocorkan credential dalam pesan kesalahan.
- Memperlakukan command kustom sebagai input tidak tepercaya hingga ditinjau pengguna.

Perlindungan terhadap perubahan filesystem bersamaan harus diterapkan sejauh dapat dilakukan secara portabel. Jika implementasi tidak dapat memastikan target file aman untuk ditulis, operasi harus gagal secara tertutup (*fail closed*) alih-alih melanjutkan secara spekulatif.

### NFR-06: Usability

CLI menyediakan help, pesan kesalahan yang dapat ditindaklanjuti, ringkasan file yang dibuat atau diganti, dan ringkasan asumsi/default penting. Semua operasi inti dapat dijalankan tanpa TUI.

### NFR-07: Testing dan Coverage

Paket core harus memiliki unit tests dan integration tests. Target awal statement coverage adalah minimal 80% untuk paket core:

- `config`
- `detector`
- `generator`
- `output`
- `validator`
- `filesystem`

Coverage bukan satu-satunya syarat rilis. Skenario keamanan, rollback, konflik, dry-run, parsing konfigurasi, serta determinisme wajib diuji secara eksplisit.

### NFR-08: Distribusi

Build dan release diotomatisasi menggunakan GitHub Actions dan GoReleaser. Setiap rilis harus memiliki versi yang dapat diidentifikasi dan artefak yang sesuai dengan platform yang dinyatakan didukung serta diuji.

---

## 6. Arsitektur Sistem

### 6.1 Tech Stack

| Komponen | Teknologi |
|---|---|
| Bahasa | Go |
| CLI | Cobra |
| TUI | Bubble Tea |
| Template engine | `text/template` |
| Template bundling | `embed.FS` |
| Unit testing | Standard library `testing` |
| CI | GitHub Actions |
| Release | GoReleaser |
| YAML parser | Library YAML Go yang teruji |

### 6.2 Struktur Repositori

```text
devopsgen/
├── cmd/
│   └── devopsgen/
│       └── main.go
├── internal/
│   ├── cli/
│   ├── wizard/
│   ├── detector/
│   ├── config/
│   ├── templates/
│   ├── generator/
│   ├── output/
│   ├── validator/
│   └── filesystem/
├── templates/
│   ├── docker/
│   ├── github-actions/
│   └── kubernetes/
├── tests/
├── docs/
├── go.mod
├── Makefile
├── README.md
└── LICENSE
```

Struktur tersebut merupakan baseline. Implementasi boleh menambahkan subpaket bila diperlukan, dengan tetap mempertahankan pemisahan tanggung jawab.

### 6.3 Tanggung Jawab Modul

- `cli`: registrasi command, flags, help, pemetaan hasil ke exit code.
- `wizard`: pengumpulan dan konfirmasi konfigurasi melalui TUI.
- `detector`: analisis file indikator stack tanpa mengubah proyek atau menjalankan script.
- `config`: model, enum, resolusi nilai, dan validasi konfigurasi.
- `templates`: pemuatan template bawaan dari `embed.FS`.
- `generator`: rendering template dan pembentukan artefak.
- `output`: representasi artefak di memori.
- `validator`: validasi sintaks dan konsistensi artefak.
- `filesystem`: validasi path, rencana perubahan, konflik, dry-run, backup, commit, dan rollback.

### 6.4 Kontrak API Core Engine

Contoh kontrak konseptual berikut menggambarkan batas tanggung jawab. Implementasi boleh menyesuaikan nama atau pembagian interface selama perilaku dan pemisahan dependensi tetap dipertahankan.

```go
package config

type StackType string
type AppType string

type Config struct {
    Stack      StackType
    AppType    AppType
    CI         string
    Registry   string
    Image      string
    Port       int
    BuildCmd   string
    TestCmd    string
    StartCmd   string
    Tag        string
    K8s        K8sOptions
    OutputDir  string
}
```

`Config` juga harus mencakup opsi yang dibutuhkan untuk Ingress, health path, dan konfigurasi khusus lain yang benar-benar dipakai oleh template MVP. Field wajib harus tervalidasi sebelum proses generator dijalankan. Contoh model opsi Kubernetes dan opsi operasi adalah sebagai berikut:

```go
type K8sOptions struct {
    Enabled    bool
    Ingress    bool
    IngressHost string
    HealthPath string
}

type GenerationOptions struct {
    Force  bool
    DryRun bool
}
```

`Force` dan `DryRun` merupakan kebijakan operasi filesystem, bukan karakteristik aplikasi. Implementasi dapat meneruskannya melalui `GenerationOptions` ke modul filesystem.

```go
package output

type Artifact struct {
    Path    string
    Content []byte
}
```

`Artifact.Path` harus berupa path relatif yang bersih dan tidak boleh mengarah ke luar root target. Implementasi boleh menambahkan metadata seperti permission yang diinginkan jika dibutuhkan.

```go
package generator

type Generator interface {
    Generate(cfg config.Config) ([]output.Artifact, error)
}
```

```go
package validator

type Validator interface {
    Validate(artifacts []output.Artifact) error
}
```

Modul filesystem harus memiliki tahapan terpisah untuk membuat rencana, memeriksa konflik, menjalankan dry-run, menyiapkan backup, commit, dan rollback. Operasi filesystem harus dapat diganti atau diinjeksi pada pengujian agar kegagalan I/O dapat disimulasikan.

### 6.5 Prinsip Ketergantungan

CLI dan TUI memakai model konfigurasi serta core engine yang sama. Core engine tidak boleh mengimpor Cobra atau Bubble Tea. Template standar harus tersedia di binary dan tidak bergantung pada folder template eksternal ketika aplikasi berjalan.

Detektor tidak boleh menjalankan perintah proyek. Core engine tidak boleh mengakses jaringan atau credential untuk rendering dan validasi internal.

### 6.6 Alur Pemrosesan

```text
CLI / TUI
   |
   v
Parsing Flags + Validasi Nilai Eksplisit  -> exit 2
   |
   v
Deteksi Stack bila Dibutuhkan             -> exit 5 bila ambigu/tidak didukung
   |
   v
Penyelesaian + Validasi Config Lengkap    -> exit 2
   |
   v
Rendering Seluruh Artefak di Memori
   |
   v
Validasi Seluruh Output                   -> exit 4
   |
   v
Validasi Path + Rencana + Konflik         -> exit 3 bila konflik tanpa izin
   |
   v
Backup + Commit Filesystem                -> exit 1 bila I/O/rollback gagal
   |
   v
Ringkasan Hasil                           -> exit 0
```

Dry-run berhenti setelah rencana dan konflik dilaporkan; tidak memasuki tahap backup atau commit. Error pada suatu tahap menghentikan alur sesuai aturan FR-16.

---

## 7. Keamanan dan Keandalan

### 7.1 Pengelolaan Secret

Credential hanya boleh direferensikan melalui mekanisme yang sesuai, seperti GitHub Actions Secrets atau `GITHUB_TOKEN`. Secret tidak boleh ditanamkan ke file hasil generasi, default konfigurasi, log, atau pesan kesalahan.

### 7.2 Path Traversal dan Symlink

Semua path artefak harus relatif terhadap root target. Sistem harus menolak path absolut yang tidak diizinkan, segmen traversal yang keluar dari target, serta resolusi symlink yang memungkinkan penulisan di luar target.

Pemeriksaan harus dilakukan saat menyusun rencana dan diulang segera sebelum operasi tulis. Implementasi harus mempertimbangkan perubahan filesystem selama transaksi sejauh dapat dilakukan secara portabel. Symlink yang menjadi target artefak harus ditolak, termasuk ketika `--force` digunakan.

Jika keamanan jalur tidak dapat dipastikan, proses harus gagal tanpa menulis file. Dokumentasi harus menyatakan batas perlindungan terhadap kondisi filesystem yang berubah bersamaan (*time-of-check to time-of-use*).

### 7.3 Penanganan Konflik

Semua konflik diperiksa sebelum commit. Tanpa `--force`, file yang konflik tidak boleh diubah. Dengan `--force`, hanya file reguler yang telah dicadangkan yang boleh diganti. Direktori, symlink, dan tujuan yang tidak aman tidak boleh ditimpa.

### 7.4 Transaksi Filesystem

Transaksi melacak file baru, file yang diganti, backup, staging, dan direktori yang dibuat. Rollback hanya memengaruhi perubahan transaksi. Jika pemulihan gagal, aplikasi melaporkan operasi yang gagal beserta jalur file yang perlu diperiksa secara manual.

### 7.5 Eksekusi Eksternal

DevOpsGen tidak boleh menjalankan `docker push`, `kubectl apply`, atau perintah deployment sebagai efek samping proses generasi. Validasi eksternal bersifat opsional, harus dipisahkan dari proses penulisan, dan hasilnya dilaporkan transparan.

---

## 8. Kriteria Penerimaan dan Pengujian

### AC-01: Inisialisasi

Perintah `init` memproses konfigurasi lengkap dan valid serta menghasilkan artefak sesuai stack dan opsi yang dipilih. Konfigurasi tidak lengkap harus menyebut parameter yang belum tersedia.

### AC-02: Mode Non-interaktif

Pengujian membuktikan proses tidak menunggu input stdin ketika parameter wajib belum tersedia, stack ambigu, atau konflik file ditemukan.

### AC-03: Deteksi Stack

Direktori dengan `go.mod` terdeteksi sebagai kandidat Go. Direktori dengan `package.json` dianalisis berdasarkan dependency, scripts, lockfile, dan indikator framework. Konflik indikator dilaporkan sebagai ambiguitas.

### AC-04: Deterministic Output

Eksekusi berulang dengan versi binary dan konfigurasi yang sama menghasilkan isi dan urutan artefak identik.

### AC-05: No-overwrite

Generasi tanpa `--force` tidak mengubah file konflik. Symlink dan direktori target ditolak meskipun `--force` diberikan.

### AC-06: Dry-run

Dry-run tidak membuat atau mengubah file, direktori, permission, timestamp, maupun metadata proyek pada direktori target. Dry-run tetap melakukan rendering, validasi, dan pemeriksaan konflik.

### AC-07: Rollback

Kegagalan I/O yang disimulasikan memicu best-effort rollback. File baru dibersihkan dan file lama dipulihkan jika memungkinkan. Kegagalan rollback harus dilaporkan terpisah dan menghasilkan exit code `1`.

### AC-08: Validasi Kubernetes

Validator mendeteksi YAML tidak valid, field wajib yang hilang, image kosong, selector-label tidak konsisten, relasi Service/Ingress yang salah, dan port yang tidak konsisten dengan konfigurasi.

### AC-09: Input Safety

Pengujian mencakup port di luar rentang, nama resource/image tidak valid, `app-type` di luar enum, path traversal, path absolut yang tidak diizinkan, symlink yang keluar dari target, dan command kustom berisi newline atau karakter kontrol.

### AC-10: Credential-free Testing

Semua unit dan integration tests berjalan tanpa credential registry atau akses cloud sungguhan.

### AC-11: Coverage

Coverage paket core ditargetkan minimal 80%, disertai pengujian khusus untuk skenario keamanan, kegagalan, konflik, dan rollback.

### AC-12: Cross-platform

Pipeline membangun target platform yang ditetapkan dan menjalankan pengujian pada runner native jika tersedia. Status platform yang hanya di-cross-compile harus dinyatakan secara berbeda dari platform yang telah diuji native.

### AC-13: Benchmark

Benchmark core menggunakan maksimal 20 artefak dengan total output maksimal 256 KiB dan target kurang dari satu detik, tanpa jaringan atau proses eksternal. Metodologi serta mesin referensi didokumentasikan.

### AC-14: Registry Workflow

Workflow menerbitkan tag commit SHA, menggunakan autentikasi sesuai registry, menerapkan permission minimum, dan tidak menyimpan credential plaintext.

### AC-15: Exit Code dan Short-circuit

Pengujian membuktikan aturan FR-16. Contoh:

- Input eksplisit tidak valid menghasilkan `2` dan tahap deteksi/generasi tidak dijalankan.
- Stack yang benar-benar ambigu menghasilkan `5` ketika tidak ada input eksplisit yang lebih awal invalid.
- Output yang tidak valid menghasilkan `4` sebelum pemeriksaan konflik.
- Konflik tanpa `--force` menghasilkan `3` setelah input serta artefak lolos validasi.
- Kegagalan I/O atau rollback menghasilkan `1`.

### AC-16: Test Stage Transparan

Jika test command tidak dikonfigurasi, workflow menampilkan notice atau summary bahwa test dilewati dan tidak mengeluarkan pesan bahwa tes berhasil. Jika test command tersedia dan gagal, proses publikasi image tidak boleh dijalankan.

### AC-17: Port dan Health Probe

`EXPOSE`, `containerPort`, dan Service `targetPort` harus mencerminkan konfigurasi port. Probe HTTP hanya dibuat jika health path yang valid dikonfigurasi atau terdeteksi dari indikator eksplisit.

---

## 9. Roadmap Implementasi

### Fase 1: Project Foundation

- Inisialisasi module Go.
- Setup Cobra dan command root.
- Buat struktur paket dasar.
- Tambahkan Makefile.
- Setup GitHub Actions untuk pengujian dasar.

### Fase 2: Config dan Output Models

- Definisikan model konfigurasi dan enum AppType.
- Implementasikan parsing dan validasi input.
- Definisikan representasi artefak di memori.
- Tambahkan unit tests dan aturan exit code.

### Fase 3: Template Engine

- Implementasikan `embed.FS`.
- Buat Dockerfile Go dan Node.js.
- Buat workflow GitHub Actions untuk Docker Hub dan GHCR.
- Buat manifest Kubernetes.
- Uji determinisme dan validitas output.

### Fase 4: Filesystem Safety

- Implementasikan validasi path dan deteksi konflik.
- Tambahkan dry-run dan force.
- Implementasikan backup, commit, dan best-effort rollback.
- Uji kegagalan I/O, symlink, serta keamanan path.

### Fase 5: Detect dan Validate

- Implementasikan deteksi `go.mod`.
- Implementasikan analisis `package.json`, scripts, lockfile, dan framework.
- Buat validasi YAML dan konsistensi Kubernetes/workflow.
- Tambahkan status validasi eksternal.

### Fase 6: CLI dan TUI

- Implementasikan `init`, `detect`, `validate`, `list`, dan `version` beserta flags.
- Hubungkan CLI ke core engine.
- Implementasikan wizard Bubble Tea.
- Terapkan perilaku non-interaktif dan pemetaan exit code.
- Tambahkan help dan pesan kesalahan konsisten.

### Fase 7: Release Engineering

- Lengkapi pipeline CI.
- Integrasikan GoReleaser.
- Bangun binary lintas platform.
- Dokumentasikan instalasi, penggunaan, asumsi, serta batasan.
- Terapkan versioning dan changelog.

### Fase 8: Ekspansi Setelah MVP

Setelah seluruh acceptance criteria MVP terpenuhi, fitur berikut dapat dipertimbangkan:

- Python.
- PHP/Laravel.
- Helm.
- Argo CD atau Flux.
- Provider CI/CD tambahan.
- Template eksternal.

---

## 10. Risiko dan Mitigasi

| Risiko | Mitigasi |
|---|---|
| Deteksi stack ambigu | Minta konfigurasi eksplisit dan jangan menebak. |
| Dockerfile tidak sesuai aplikasi | Validasi asumsi runtime, uji per template, dan dokumentasikan keterbatasan deteksi statis. |
| Credential terekspos | Gunakan secret references, permission minimum, dan uji artefak/log. |
| File rusak ketika proses gagal | Gunakan rencana perubahan, backup, pencatatan transaksi, dan best-effort rollback. |
| Symlink atau filesystem berubah saat operasi | Validasi path berulang, tolak target tidak aman, dan fail closed jika tidak dapat dipastikan. |
| Manifest tidak cocok dengan cluster | Lakukan validasi internal, dokumentasikan prasyarat, dan jangan menjalankan apply otomatis. |
| Port container berbeda dari listener aplikasi | Nyatakan batasan `EXPOSE`/`containerPort`, validasi konfigurasi, dan dokumentasikan bahwa aplikasi harus bind ke port tersebut. |
| Test stage tidak tersedia | Nyatakan tahap dilewati secara eksplisit; jangan melaporkan test berhasil. |
| Perbedaan platform | CI multi-platform dan pengujian native bila tersedia. |
| Scope creep | Tegakkan batas MVP dan pindahkan ekspansi ke fase berikutnya. |
| Coverage tinggi tetapi kualitas rendah | Uji skenario keamanan, regresi, determinisme, dan kegagalan secara eksplisit. |
| Template berubah tanpa pelacakan | Ikat template ke versi binary serta changelog. |

---

## 11. Definition of Done

MVP dinyatakan siap dirilis apabila:

1. Seluruh command wajib tersedia dan menyediakan help.
2. Mode interaktif dan non-interaktif mengikuti spesifikasi.
3. Stack Go dan Node.js yang didukung menghasilkan artefak sesuai konfigurasi.
4. No-overwrite, dry-run, dan force telah diuji.
5. Backup dan best-effort rollback diuji melalui skenario kegagalan I/O.
6. Validasi input, YAML, struktur workflow, dan konsistensi Kubernetes berjalan.
7. Target coverage core dievaluasi dan skenario kritis diuji eksplisit.
8. Pipeline CI lulus pada target yang ditetapkan.
9. Dokumentasi penggunaan, instalasi, keamanan, asumsi, serta batasan tersedia.
10. Binary release memiliki versi yang dapat diidentifikasi.
11. Benchmark core dijalankan dan hasilnya dicatat.
12. Tidak ada credential plaintext dalam artefak yang dihasilkan.
13. Exit code dan aturan short-circuit diuji sesuai AC-15.
14. Test stage yang tidak dikonfigurasi dilaporkan sebagai dilewati, bukan berhasil.
15. Template tidak membuat klaim bahwa port, health endpoint, atau nilai resource telah diverifikasi jika bukti tidak tersedia.

---

## 12. Persetujuan Spesifikasi

Dokumen ini menjadi baseline teknis DevOpsGen versi **1.2.0** dengan status **Approved for Implementation**.

Perubahan pada scope MVP, kontrak CLI, kebijakan keamanan, arsitektur core, model konfigurasi, atau acceptance criteria harus dicatat sebagai revisi spesifikasi atau keputusan desain. Implementasi harus mengutamakan ketepatan output, keamanan filesystem, determinisme, dan kemudahan pengujian sebelum ekspansi fitur.

**Status akhir: Approved for Implementation.**
