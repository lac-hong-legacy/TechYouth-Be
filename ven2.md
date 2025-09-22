
# Tài liệu Yêu cầu Sản phẩm - Ứng dụng Ven

## Phiên bản 2.0 - MVP Mobile

  

---

  

## 1. TÓM TẮT

  

### 1.1 Tổng quan Sản phẩm

**Ven** là ứng dụng EdTech gamified dạy lịch sử Việt Nam theo phương pháp "học mà chơi", biến việc học sử thành trải nghiệm thú vị và dễ tiếp cận cho mọi lứa tuổi.

  

### 1.2 Giá trị Cốt lõi

- **Học nhanh**: Mỗi bài học chỉ 3-5 phút

- **Trực quan**: Ưu tiên hình ảnh, animation thay vì văn bản

- **Gamified**: Tích hợp sâu yếu tố game (Linh thú, cấp độ, huy hiệu)

- **Cảm xúc**: Kể chuyện với voice-over truyền cảm

  

### 1.3 Đối tượng Người dùng

1. **Học sinh - Sinh viên** (12-25 tuổi): Cần học lịch sử vui nhộn, nhanh gọn

2. **Người đi làm** (25-40 tuổi): Muốn ôn lại kiến thức trong thời gian rảnh

  

### 1.4 Phạm vi MVP

- 20 nhân vật/sự kiện lịch sử

- Vòng lặp học tập cốt lõi với gamification

- Chế độ khách với giới hạn

- Kiếm tiền từ quảng cáo

- Tính năng cộng đồng cơ bản

  

---

  

## 2. PHÂN LOẠI NGƯỜI DÙNG & QUYỀN HẠN

  

### 2.1 Người dùng Khách (Chưa đăng ký)

**Quyền hạn:**

- Được học 2 bài học đầu tiên

- Tiến trình lưu cục bộ trên thiết bị

- Không được reset tim hàng ngày

- Phải xem quảng cáo khi hết tim

- Không truy cập tính năng cộng đồng

  

**Giới hạn:**

- Sau 2 bài học → Bắt buộc đăng ký

- Không đồng bộ dữ liệu giữa các thiết bị

- Không tham gia bảng xếp hạng

  

### 2.2 Người dùng Đã đăng ký

**Quyền hạn:**

- Học không giới hạn bài học

- Đồng bộ tiến trình qua cloud

- Reset 5 tim mỗi ngày (0:00 sáng)

- Tham gia bảng xếp hạng

- Chia sẻ thành tích

- Mở khóa toàn bộ nhân vật

  

---

  

## 3. CẤU TRÚC THÔNG TIN

  

### 3.1 Cấu trúc Nội dung

```

Dòng thời gian Lịch sử Việt Nam

├── Thời Tiền Sử

│ └── [Khóa cho MVP]

├── Thời Bắc Thuộc

│ └── Hai Bà Trưng (Huyền thoại)

├── Thời Độc Lập

│ ├── Nhà Đinh

│ │ └── Đinh Bộ Lĩnh (Hiếm)

│ ├── Nhà Lý

│ │ ├── Lý Thường Kiệt (Huyền thoại)

│ │ └── [Sự kiện phụ]

│ ├── Nhà Trần

│ │ ├── Trần Hưng Đạo (Huyền thoại)

│ │ ├── Trần Quốc Toản (Phổ thông)

│ │ └── Yết Kiêu (Hiếm)

│ └── [Các triều đại khác...]

```

  

### 3.2 Cấu trúc Điều hướng

```

Ứng dụng

├── Màn hình Splash

├── Luồng Giới thiệu (Có thể bỏ qua)

├── Màn hình Chính (Bản đồ Timeline)

├── Module Học tập

├── Bộ sưu tập

├── Bảng xếp hạng

├── Hồ sơ & Cài đặt

└── Luồng Xác thực

```

  

---

  

## 4. LUỒNG NGƯỜI DÙNG CHI TIẾT

  

### 4.1 TRẢI NGHIỆM NGƯỜI DÙNG LẦN ĐẦU

  

#### Luồng A: Đường dẫn Khách (Bỏ qua Đăng ký)

```

BẮT ĐẦU

└── Màn hình Splash (2 giây)

└── Màn hình Chào mừng

├── [Bỏ qua] → Timeline Chính (Chế độ Khách)

│ └── Có thể khám phá giao diện

│ └── Có thể bắt đầu 2 bài học đầu

│ └── Sau 2 bài học → Bắt buộc hiện popup Đăng ký

│

└── [Bắt đầu] → Luồng Đăng ký

```

  

#### Luồng B: Đường dẫn Đăng ký

```

BẮT ĐẦU

└── Màn hình Splash (2 giây)

└── Màn hình Chào mừng

└── [Bắt đầu]

└── Màn hình Tùy chọn Đăng ký

├── Đăng nhập Google → Luồng OAuth

├── Đăng nhập Facebook → Luồng OAuth

└── Đăng ký Email

└── Nhập: Tên người dùng, Email, Mật khẩu

└── Xác thực Email (Tùy chọn cho MVP)

└── Thành công → Luồng Giới thiệu

LUỒNG GIỚI THIỆU:

└── Bước 1: "Chào mừng đến với Ven!"

└── Bước 2: Nhập Năm sinh

└── Bước 3: Nhận Trứng Linh thú (Theo con giáp)

└── Bước 4: Bài học Hướng dẫn

└── Trứng nở → Hiện Linh thú

└── Vào Timeline Chính

```

  

### 4.2 VÒNG LẶP HỌC TẬP CỐT LÕI

  

#### Luồng Hoàn thành Bài học

```

BẮT ĐẦU BÀI HỌC

├── Người dùng chạm vào Node trên Timeline

├── Kiểm tra Điều kiện tiên quyết

│ ├── [Đã khóa] → Hiện yêu cầu ("Hoàn thành X trước")

│ └── [Đã mở khóa] → Tiếp tục

│

├── Màn hình Tải Bài học

├── Phần 1: KỂ CHUYỆN (1-2 phút)

│ ├── Video/Animation ngắn phát

│ ├── Voice-over với phụ đề

│ ├── [Có thể bỏ qua sau 5 giây]

│ └── Tự động chuyển sang Quiz

│

├── Phần 2: TƯƠNG TÁC QUIZ (3-5 câu hỏi)

│ ├── Loại câu hỏi:

│ │ ├── Trắc nghiệm (4 lựa chọn)

│ │ ├── Sắp xếp Timeline (Kéo & Thả)

│ │ ├── Nối Nhân vật-Sự kiện

│ │ └── Điền vào chỗ trống

│ │

│ ├── Với mỗi câu hỏi:

│ │ ├── Hiện câu hỏi → Người dùng trả lời

│ │ ├── Phản hồi ngay lập tức:

│ │ │ ├── [Đúng] → Linh thú ăn mừng

│ │ │ │ └── Animation +XP → Câu tiếp theo

│ │ │ └── [Sai] → Linh thú buồn

│ │ │ ├── Mất 1 tim

│ │ │ ├── [Tim > 0] → Thử lại

│ │ │ └── [Tim = 0] → Luồng hết tim

│ │

│ └── Hoàn thành tất cả câu hỏi → Tiếp tục

│

├── Phần 3: MÀN HÌNH HOÀN THÀNH BÀI HỌC

│ ├── Hiển thị XP kiếm được

│ ├── Hiện thanh tiến trình nhân vật/sự kiện

│ ├── [Nếu đạt mốc] → Animation mở khóa

│ ├── Hiện Quảng cáo (Xen kẽ/Có thưởng)

│ └── [Tiếp tục] → Quay lại Timeline

│

└── KẾT THÚC BÀI HỌC

```

  

#### Luồng Hết Tim

```

TIM = 0

├── Popup: "Bạn đã hết tim!"

├── Tùy chọn:

│ ├── [Người dùng Khách]:

│ │ └── Phải xem quảng cáo → +3 tim

│ │ └── Tiếp tục bài học

│ │

│ └── [Người dùng Đã đăng ký]:

│ │ ├── Xem quảng cáo → +3 tim

│ │ └── Đợi reset hàng ngày (hiện đếm ngược)

│

└── [Thoát] → Quay lại Timeline

```

  

### 4.3 HỆ THỐNG MỞ KHÓA NHÂN VẬT

  

#### Luồng Mở khóa theo Độ hiếm (CẢ 3 CẤP ĐỘ ĐỀU CÓ THANH TIẾN TRÌNH)

```

NHÂN VẬT PHỔ THÔNG (1-2 bài học)

├── Hiện thanh tiến trình (0%)

├── Hoàn thành bài 1 → 50% hoặc 100%

├── [Nếu 2 bài] Hoàn thành bài 2 → 100%

├── Animation mở khóa đơn giản

├── Hiện thẻ nhân vật

└── Thêm vào bộ sưu tập

  

NHÂN VẬT HIẾM (2-3 bài học)

├── Hiện thanh tiến trình (0%)

├── Bài học 1 → 33% + Mảnh ghép 1/3

├── Bài học 2 → 66% + Mảnh ghép 2/3

├── Bài học 3 → 100% + Mảnh ghép 3/3

├── Thu thập đủ mảnh ghép

│ └── Animation mở khóa trung bình

│ └── Nhân vật xuất hiện

│ └── Thêm vào bộ sưu tập

  

NHÂN VẬT HUYỀN THOẠI (4-5 bài học)

├── Hiện thanh tiến trình (0%)

├── Mỗi bài học → +20-25% tiến trình

├── Thu thập hiện vật dọc đường

├── Hoàn thành 100%

│ └── Chuỗi mở khóa điện ảnh

│ ├── Nhạc nền hùng tráng

│ ├── Voice-over câu nói nổi tiếng

│ ├── Animation nhân vật

│ └── Thẻ cao cấp xuất hiện

```

  

### 4.4 HỆ THỐNG GAMIFICATION

  

#### Hệ thống Chuỗi ngày

```

NGƯỜI DÙNG MỞ ỨNG DỤNG

├── Kiểm tra thời gian hoạt động cuối

├── [Trong vòng 24h từ bài học cuối]:

│ └── Tiếp tục chuỗi

│ └── Cập nhật bộ đếm

│

├── [Hơn 24h]:

│ └── Reset chuỗi về 0

│ └── Hiện thông báo mất chuỗi

│

└── Hiển thị chuỗi hiện tại trên màn hình chính

```

  

#### Tiến hóa Linh thú

```

TIẾN TRÌNH LINH THÚ

├── Giai đoạn 1: Trứng (0 XP)

│ └── Hoàn thành hướng dẫn → Nở

│

├── Giai đoạn 2: Dạng baby (1-500 XP)

│ └── Hình ảnh: Nhỏ, dễ thương

│

├── Giai đoạn 3: Dạng thiếu niên (501-2000 XP)

│ └── Hình ảnh: Cỡ trung, chi tiết hơn

│

├── Giai đoạn 4: Dạng trưởng thành (2001-5000 XP)

│ └── Hình ảnh: Kích thước đầy đủ, uy nghi

│

└── Giai đoạn 5: Dạng huyền thoại (5000+ XP)

└── Hình ảnh: Hiệu ứng đặc biệt, hào quang

```

  

### 4.5 BỘ SƯU TẬP & THÀNH TÍCH

  

#### Luồng Xem Bộ sưu tập

```

PHÒNG TRƯNG BÀY BỘ SƯU TẬP

├── Tab: Nhân vật

│ ├── Tùy chọn lọc:

│ │ ├── Theo Triều đại

│ │ ├── Theo Độ hiếm

│ │ └── Đã mở khóa/Chưa mở khóa

│ │

│ └── Lưới Nhân vật

│ ├── [Chạm đã mở khóa] → Xem chi tiết

│ │ ├── Artwork đầy đủ

│ │ ├── Tiểu sử

│ │ ├── Các mốc sự kiện

│ │ ├── Câu nói nổi tiếng

│ │ └── [Chia sẻ] → Luồng chia sẻ

│ │

│ └── [Chạm chưa mở khóa] → Hiện yêu cầu

│

└── Tab: Huy hiệu

├── Danh mục:

│ ├── Cột mốc học tập

│ ├── Thành tích chuỗi ngày

│ └── Sự kiện đặc biệt

│

└── Lưới Huy hiệu

└── [Chạm] → Xem chi tiết

```

  

### 4.6 TÍNH NĂNG XÃ HỘI & CỘNG ĐỒNG

  

#### Luồng Bảng xếp hạng

```

BẢNG XẾP HẠNG

├── Mặc định: Xem theo Tuần

├── Các tab:

│ ├── Tuần

│ ├── Tháng

│ └── Mọi thời gian

│

├── Vị trí người dùng được highlight

├── Top 3 với huy hiệu đặc biệt

├── Tải thêm khi cuộn (20 người/trang)

└── [Chạm tên] → Xem hồ sơ cơ bản

```

  

#### Luồng Chia sẻ Thành tích

```

MỞ KHÓA THÀNH TÍCH

├── [Nút chia sẻ]

├── Tạo thẻ chia sẻ:

│ ├── Hình ảnh thành tích

│ ├── Tên người dùng

│ ├── Logo ứng dụng

│ └── Câu trích dẫn truyền cảm hứng

│

└── Tùy chọn chia sẻ:

├── Facebook

├── Instagram Story

├── TikTok

└── Lưu vào thư viện

```

  

### 4.7 CÀI ĐẶT & HỒ SƠ

  

#### Cấu trúc Menu Cài đặt

```

CÀI ĐẶT

├── Phần Hồ sơ

│ ├── Avatar & tên người dùng

│ ├── Cấp độ & tiến trình XP

│ ├── Chuỗi ngày hiện tại

│ ├── Tổng thành tích

│ └── Hiển thị linh thú

│

├── Quản lý Tài khoản

│ ├── Đổi mật khẩu

│ ├── Tài khoản liên kết

│ ├── Tùy chọn email

│ └── Đăng xuất

│

├── Cài đặt Ứng dụng

│ ├── Âm thanh (Bật/Tắt)

│ ├── Nhạc nền (Bật/Tắt)

│ ├── Thông báo (Bật/Tắt)

│ └── Ngôn ngữ (Chỉ Tiếng Việt cho MVP)

│

└── Thông tin

├── Điều khoản Dịch vụ

├── Chính sách Bảo mật

└── Thông tin phiên bản

```

  

---

  

## 5. HỆ THỐNG NỘI DUNG ĐA TẦNG

  

### 5.1 Tổng quan Kiến trúc Nội dung

  

Hệ thống nội dung đa tầng được thiết kế để phục vụ nhiều nhóm người dùng với các mức độ quan tâm khác nhau, từ người học casual đến những người đam mê lịch sử sâu sắc. Mỗi nhân vật/sự kiện đều có tiềm năng mở rộng lên 3 tầng, tùy thuộc vào lượng dữ liệu lịch sử có sẵn.

  

```

CẤU TRÚC NỘI DUNG ĐA TẦNG

│

├── TẦNG 1: Trải nghiệm Cốt lõi (100% người dùng)

│ └── Bắt buộc cho mọi nhân vật/sự kiện

│

├── TẦNG 2: Khám phá Chuyên sâu (30-40% người dùng)

│ └── Tùy chọn, dành cho người muốn tìm hiểu thêm

│

└── TẦNG 3: Thử thách Sử gia (5-10% người dùng)

└── Dành cho người đam mê, muốn thử thách kiến thức

```

  

### 5.2 CHI TIẾT TỪNG TẦNG

  

#### TẦNG 1: Trải nghiệm Cốt lõi (Core Experience)

  

**Mục đích:**

- Cung cấp kiến thức nền tảng về nhân vật/sự kiện

- Tạo trải nghiệm vui vẻ, không áp lực

- Xây dựng thói quen học tập hàng ngày

  

**Cấu trúc Bài học:**

```

MỖI BÀI HỌC TẦNG 1

├── PHẦN 1: Hook & Storytelling (30%)

│ ├── Opening Hook (15 giây)

│ │ └── Câu hỏi/tình huống gây tò mò

│ ├── Main Story (60-90 giây)

│ │ ├── Animation/Illustrations

│ │ ├── Voice-over chuyên nghiệp

│ │ └── Subtitle tùy chọn

│ └── Transition to Quiz

│

├── PHẦN 2: Interactive Quiz (50%)

│ ├── Câu hỏi 1: Warm-up (Dễ)

│ ├── Câu hỏi 2-3: Core content (Trung bình)

│ ├── Câu hỏi 4-5: Challenge (Khó hơn chút)

│ └── Instant feedback cho mỗi câu

│

└── PHẦN 3: Reward & Progress (20%)

├── XP earned animation

├── Progress bar update

├── Unlock preview (nếu có)

└── Motivational message

```

  

**Loại Nội dung:**

- Sự kiện chính, cột mốc quan trọng

- Thông tin cơ bản về nhân vật

- Câu nói/hành động nổi tiếng

- Context lịch sử đơn giản

  

**Thời lượng:** 3-5 phút/bài

  

#### TẦNG 2: Khám phá Chuyên sâu (Deep Dive)

  

**Mục đích:**

- Cung cấp bối cảnh và chiều sâu lịch sử

- Kết nối với tư liệu thực tế

- Tạo cảm giác "khám phá kho báu"

  

**Điều kiện Mở khóa:**

- Hoàn thành 100% bài học Tầng 1 của nhân vật/sự kiện

- Thanh "Khám phá thêm" xuất hiện

  

**Cấu trúc Nội dung:**

  

```

TẦNG 2 - KHÁM PHÁ CHUYÊN SÂU

│

├── MODULE 1: Hiện vật & Tư liệu (Artifacts Gallery)

│ ├── Hình ảnh độ phân giải cao

│ ├── Khả năng zoom/pan

│ ├── Thông tin chi tiết khi tap

│ └── Nguồn/Bảo tàng lưu giữ

│

├── MODULE 2: Góc nhìn Đa chiều (Multiple Perspectives)

│ ├── Tab 1: Góc nhìn Việt Nam

│ ├── Tab 2: Góc nhìn đối phương

│ ├── Tab 3: Nhận xét từ sử gia

│ └── Tab 4: Tác động đến dân chúng

│

├── MODULE 3: Timeline Mở rộng (Extended Timeline)

│ ├── Sự kiện trước đó (nguyên nhân)

│ ├── Diễn biến chi tiết

│ ├── Hậu quả ngắn hạn

│ └── Ảnh hưởng dài hạn

│

└── MODULE 4: Kết nối Liên quan (Related Connections)

├── Nhân vật liên quan

├── Sự kiện tương tự

├── So sánh quốc tế

└── Di sản còn lại

```

  

**Ví dụ Chi tiết - Trần Hưng Đạo:**

  

1. **Hiện vật & Tư liệu:**

- Ảnh chụp bản gốc "Hịch tướng sĩ"

- Bản đồ chiến lược Vạn Kiếp

- Hình ảnh cọc Bạch Đằng được khai quật

- Tem thư, tiền giấy có hình Trần Hưng Đạo

  

2. **Góc nhìn Đa chiều:**

- **Việt Nam**: Trích "Đại Việt sử ký toàn thư"

- **Nguyên Mông**: Ghi chép từ "Nguyên sử"

- **Sử gia hiện đại**: Phân tích chiến thuật

- **Dân gian**: Truyền thuyết về thần Trần

  

3. **Timeline Mở rộng:**

- 1257: Lần xâm lược đầu tiên

- 1284: Chuẩn bị kháng chiến

- 1285: Lần xâm lược thứ hai

- 1287-1288: Lần xâm lược thứ ba

- Sau 1288: Xây dựng hòa bình

  

4. **Kết nối Liên quan:**

- So sánh với Ngô Quyền (Bạch Đằng 938)

- Ảnh hưởng đến Lê Lợi, Quang Trung

- Tương đồng với các chiến lược hải quân thế giới

  

**Interface Design:**

- Dạng scrollable cards hoặc tabs

- Hình ảnh có thể phóng to full screen

- Text có thể điều chỉnh size

- Bookmark để xem lại sau

  

#### TẦNG 3: Thử thách Sử gia (Scholar Mode)

  

**Mục đích:**

- Test kiến thức chuyên sâu

- Tạo thành tựu đặc biệt cho hardcore users

- Xây dựng cộng đồng "sử gia" trong app

  

**Điều kiện Mở khóa:**

- Đã xem qua Tầng 2

- Nút "Thử thách Sử gia" xuất hiện

  

**Cấu trúc Thử thách:**

  

```

TẦNG 3 - THỬ THÁCH SỬ GIA

│

├── PHẦN 1: Quiz Chuyên sâu (Expert Quiz)

│ ├── 10 câu hỏi khó

│ ├── Thời gian giới hạn (30s/câu)

│ ├── Không được dùng tim

│ └── Phải đạt 80% để pass

│

├── PHẦN 2: Mini-game Tương tác (Interactive Challenge)

│ ├── Xếp đặt chiến thuật trên bản đồ

│ ├── Ghép câu từ văn bản gốc

│ ├── Nhận diện hiện vật

│ └── Sắp xếp timeline phức tạp

│

└── PHẦN 3: Essay Challenge (Tùy chọn)

├── Câu hỏi mở

├── Trả lời 100-200 từ

├── AI hoặc cộng đồng đánh giá

└── Nhận badge đặc biệt

```

  

**Loại Câu hỏi Mẫu:**

  

1. **Chi tiết Lịch sử:**

- "Trận Vân Đồn diễn ra vào ngày nào?"

- "Tên thật của Phạm Ngũ Lão là gì?"

- "Có bao nhiêu vạn quân trong lần xâm lược thứ 3?"

  

2. **Phân tích Chiến thuật:**

- Đặt cọc đúng vị trí trên bản đồ Bạch Đằng

- Sắp xếp trình tự rút lui của quân Trần

- Chọn thời điểm thủy triều phù hợp

  

3. **Văn bản Gốc:**

- Điền từ còn thiếu trong "Hịch tướng sĩ"

- Ghép câu với tác giả (vua, tướng)

- Dịch nghĩa câu Hán Việt

  

**Phần thưởng Đặc biệt:**

- Badge "Sử gia" (Bronze/Silver/Gold)

- Title hiếm: "Người giữ lửa", "Bậc thầy lịch sử"

- Unlock avatar frame đặc biệt

- Tên lên Hall of Fame

  

### 5.3 LOGIC HIỂN THỊ & ĐIỀU KIỆN

  

#### Quy tắc Hiển thị Nội dung

  

```

DECISION TREE CHO NỘI DUNG ĐA TẦNG

│

├── Nhân vật/Sự kiện mới

│ └── Chỉ hiện Tầng 1

│

├── Hoàn thành Tầng 1 (100%)

│ ├── [Có data Tầng 2] → Hiện nút "Khám phá thêm"

│ └── [Không có data] → Chỉ hiện "Hoàn thành"

│

├── Đã xem Tầng 2

│ ├── [Có data Tầng 3] → Hiện nút "Thử thách"

│ └── [Không có data] → Kết thúc

│

└── Special Cases

├── Event đặc biệt → Mở khóa tạm thời

└── Premium user → Early access

```

  

#### Tracking Progress

  

```json

{

"nhanVatId": "tran_hung_dao",

"tienTrinhTang": {

"tang1": {

"daHoanThanh": true,

"baiHoc": [1,2,3,4,5],

"tongXP": 250,

"thoiGianHoanThanh": "2024-01-15"

},

"tang2": {

"daXem": true,

"modules": {

"hienVat": true,

"gocNhin": true,

"timeline": false,

"ketNoi": false

},

"phanTramHoanThanh": 50

},

"tang3": {

"daThach": true,

"diemQuiz": 85,

"miniGamePass": true,

"badgeNhan": "bronze_historian"

}

}

}

```

  

### 5.4 VÍ DỤ CHI TIẾT CHO TỪNG LOẠI NHÂN VẬT

  

#### A. Nhân vật LEGENDARY - Đầy đủ 3 tầng

  

**Ví dụ: Hai Bà Trưng**

  

**TẦNG 1 - Core (5 bài học):**

1. Bối cảnh Bắc thuộc & áp bức

2. Hai Bà Trưng và lý do khởi nghĩa

3. Chiến thắng & lập nước

4. Cai trị và xây dựng

5. Kết cục bi tráng & di sản

  

**TẦNG 2 - Deep Dive:**

- **Hiện vật**: Đền thờ, trống đồng Đông Sơn, tượng Hai Bà

- **Góc nhìn**:

- Sử Việt: Anh hùng dân tộc

- Sử Trung: "Loạn Trưng Trắc"

- Sử gia: Phân tích vai trò phụ nữ

- **Timeline**: 40 TCN - 43 SCN chi tiết

- **Kết nối**: Ảnh hưởng đến các cuộc khởi nghĩa sau

  

**TẦNG 3 - Scholar:**

- Quiz: Tên 36 thành trì, các tướng nữ

- Mini-game: Chiến thuật voi chiến

- Essay: "Vai trò phụ nữ trong lịch sử VN"

  

#### B. Nhân vật RARE - Có 2-3 tầng

  

**Ví dụ: Yết Kiêu**

  

**TẦNG 1 - Core (3 bài học):**

1. Sức mạnh phi thường từ lông trâu thần

2. Một mình phá giặc dưới nước

3. Bị bắt và thoát hiểm

  

**TẦNG 2 - Deep Dive:**

- **Hiện vật**: Đền thờ ở Hải Phòng

- **Góc nhìn**: Truyền thuyết dân gian vs sử sách

- **Kết nối**: Các anh hùng "siêu nhân" VN

  

**TẦNG 3:**  [Không đủ data - không hiển thị]

  

#### C. Nhân vật COMMON - Thường chỉ 1-2 tầng

  

**Ví dụ: Trần Quốc Toản**

  

**TẦNG 1 - Core (1 bài học):**

- Câu chuyện bóp nát quả cam

  

**TẦNG 2 - Deep Dive (nếu có data):**

- Gia thế và vai trò trong triều Trần

- Những trận đánh khác

- Cái chết và tang lễ

  

**TẦNG 3:**  [Không có]

  

### 5.5 GUIDELINES CHO CONTENT TEAM

  

#### Tiêu chí Quyết định Số Tầng

  

**Để có TẦNG 2, cần:**

- Ít nhất 3-5 tư liệu hình ảnh chất lượng

- 2+ góc nhìn khác nhau từ sử liệu

- Đủ thông tin cho timeline mở rộng

- Có kết nối với nhân vật/sự kiện khác

  

**Để có TẦNG 3, cần:**

- Đủ chi tiết cho 10+ câu quiz khó

- Có yếu tố tương tác (bản đồ, chiến thuật)

- Data cho mini-game có ý nghĩa

- Cộng đồng quan tâm (tracking từ Tầng 2)

  

#### Content Quality Standards

  

1. **Tầng 1**:

- Phải chính xác 100% về sự kiện chính

- Voice-over rõ ràng, truyền cảm

- Quiz cân bằng độ khó

  

2. **Tầng 2**:

- Tư liệu phải có nguồn rõ ràng

- Đa góc nhìn nhưng không gây tranh cãi

- Hình ảnh min 1080p

  

3. **Tầng 3**:

- Câu hỏi phải có trong sách giáo khoa hoặc sử liệu

- Không được có câu hỏi "mẹo"

- Essay topics phải mở và thú vị

## 6. ĐẶC TẢ NỘI DUNG

  

### 6.1 Thành phần Bài học

  

#### Đặc tả Storytelling

- **Thời lượng**: 60-120 giây

- **Định dạng**: Minh họa hoạt hình với pan/zoom

- **Âm thanh**: Voice-over chuyên nghiệp (Tiếng Việt)

- **Phụ đề**: Tùy chọn, mặc định bật

- **Bỏ qua**: Có thể sau 5 giây

  

#### Đặc tả Quiz

- **Số câu hỏi/bài**: 3-5

- **Giới hạn thời gian**: Không có cho MVP

- **Thử lại**: Không giới hạn (tốn tim)

- **Phản hồi**: Ngay lập tức sau mỗi câu trả lời

  

#### Gói Nội dung cho mỗi Nhân vật

```

GÓI NHÂN VẬT

├── Dữ liệu Hồ sơ

│ ├── Tên, danh xưng, triều đại

│ ├── Năm sinh-mất

│ ├── 2-3 thành tựu chính

│ └── 1 câu nói nổi tiếng

│

├── Tài sản Hình ảnh

│ ├── Artwork nhân vật (3 phiên bản theo độ hiếm)

│ ├── Minh họa câu chuyện (5-10 ảnh/nhân vật)

│ └── Icon cho node timeline

│

├── Tài sản Âm thanh

│ ├── File voice-over (MP3)

│ └── Nhạc nền (thư viện chung)

│

└── Nội dung Học tập

├── Kịch bản câu chuyện (100-200 từ/bài)

├── Câu hỏi quiz (10-15 câu tổng)

└── Nội dung chuyên sâu (văn bản tùy chọn)

```

  

### 6.2 Quy tắc Mở khóa Nội dung

  

#### Tuyến tính vs Phi tuyến

- **Trong Triều đại**: Phi tuyến (chọn bất kỳ node có sẵn)

- **Giữa các Triều đại**: Bán tuyến tính (mở khóa 70% để tiếp tục)

- **Sự kiện Đặc biệt**: Yêu cầu điều kiện tiên quyết cụ thể

  

#### Điều kiện Mở khóa

1. **Tiêu chuẩn**: Hoàn thành bài học trước

2. **Cột mốc**: Hoàn thành X bài học tổng

3. **Chuỗi ngày**: Duy trì chuỗi Y ngày

4. **Đặc biệt**: Sự kiện giới hạn thời gian

  

---

  

## 7. LUỒNG KIẾM TIỀN

  

### 7.1 Vị trí Quảng cáo

  

#### Sau khi Hoàn thành Bài học

```

HOÀN THÀNH BÀI HỌC

├── Hiện màn hình phần thưởng (tối thiểu 3 giây)

├── Tải quảng cáo ngầm

├── [Tự động hiện quảng cáo]

│ ├── Quảng cáo xen kẽ (5-30 giây)

│ └── [X] đóng sau đếm ngược

│

└── Quay lại timeline

```

  

#### Nạp lại Tim

```

HẾT TIM

├── Popup xuất hiện

├── Nút [Xem Quảng cáo]

├── Tải video có thưởng

├── Người dùng xem hết quảng cáo (30 giây)

├── Thưởng: +3 tim

└── Tiếp tục chơi

```

  

### 7.2 Kiếm tiền Tương lai (Sau MVP)

- Gói premium (không quảng cáo, tim vô hạn)

- Mua gói nhân vật

- Vật phẩm trang trí cho linh thú

- Vật phẩm tăng cường (XP x2, bảo vệ chuỗi)

  

---

  

## 8. QUẢN LÝ TRẠNG THÁI

  

### 8.1 Trạng thái Người dùng

  

#### Dữ liệu Người dùng Khách (Lưu trữ Cục bộ)

```json

{

"laKhach": true,

"baiHocHoanThanh": [id1, id2],

"timHienTai": 3,

"tongXP": 150,

"giaĐoanLinhThu": 1,

"thoiGianHoatDongCuoi": timestamp,

"idThietBi": "uuid"

}

```

  

#### Dữ liệu Người dùng Đã đăng ký (Đồng bộ Cloud)

```json

{

"idNguoiDung": "uuid",

"hoSo": {

"tenNguoiDung": "string",

"email": "string",

"namSinh": 1995,

"conGiap": "heo",

"ngayThamGia": timestamp

},

"tienTrinh": {

"baiHocHoanThanh": [],

"nhanVatDaMoKhoa": [],

"xpHienTai": 0,

"capDo": 1,

"timHienTai": 5,

"timToiDa": 5,

"chuoiNgay": 0,

"chuoiDaiNhat": 0,

"thoiGianHoatDongCuoi": timestamp

},

"linhThu": {

"loai": "rong",

"giaiDoan": 1,

"xpDenGiaiDoanKe": 500

},

"thanhTich": [],

"caiDat": {

"amThanh": true,

"nhacNen": true,

"thongBao": true

}

}

```

  

### 8.2 Trạng thái Bài học

  

```

TRANG_THAI_BAI_HOC

├── DA_KHOA (Chưa đủ điều kiện)

├── CO_THE_CHOI (Có thể bắt đầu)

├── DANG_CHOI (Đang chơi)

├── HOAN_THANH (Đã xong, có thể chơi lại)

└── HOAN_HAO (Tất cả câu hỏi đúng)

```

  

### 8.3 Trạng thái Mạng & Lỗi

  

#### Chế độ Offline

- **Khách**: Đầy đủ chức năng với lưu trữ cục bộ

- **Đã đăng ký**: Chế độ chỉ đọc, xếp hàng hành động để đồng bộ

  

#### Xử lý Lỗi

```

TRANG_THAI_LOI

├── KHONG_CO_MANG

│ └── Hiện banner chế độ offline

│

├── DONG_BO_THAT_BAI

│ └── Thử lại tự động (3 lần)

│

├── TAI_QUANG_CAO_THAT_BAI

│ └── Bỏ qua quảng cáo, tiếp tục luồng

│

└── TAI_NOI_DUNG_THAT_BAI

└── Hiện nút thử lại

```

  

---

  

## 9. THÔNG BÁO

  

### 9.1 Loại Thông báo Đẩy

  

#### Nhắc nhở Hàng ngày

- **Kích hoạt**: 20:00 nếu chưa học hôm nay

- **Nội dung**: "Đừng để mất chuỗi ngày! Linh thú đang đợi bạn 🐉"

  

#### Cảnh báo Chuỗi ngày

- **Kích hoạt**: 21:00 nếu chuỗi > 3 ngày

- **Nội dung**: "Chuỗi [X] ngày sắp mất! Học ngay để duy trì 🔥"

  

#### Mở khóa Thành tích

- **Kích hoạt**: Khi mở khóa hiếm/huyền thoại

- **Nội dung**: "Chúc mừng! Bạn đã mở khóa [Tên Nhân vật] 🎉"

  

### 9.2 Thông báo Trong ứng dụng

  

#### Mẹo Hướng dẫn

- Hiện một lần cho mỗi tính năng khi dùng lần đầu

- Có thể tắt, không hiện lại

  

#### Ăn mừng Cột mốc

- Animation lên cấp

- Cột mốc chuỗi ngày (7, 30, 100 ngày)

- Cột mốc bộ sưu tập

  

---

  

## 10. TRƯỜNG HỢP ĐẶC BIỆT & XỬ LÝ LỖI

  

### 10.1 Luồng Người dùng Quan trọng

  

#### Lỗi Đăng ký

```

EMAIL_DA_TON_TAI

└── "Email đã được sử dụng. Đăng nhập?"

├── [Đăng nhập] → Luồng đăng nhập

└── [Thử email khác] → Ở lại

  

LOI_MANG

└── "Không thể kết nối. Thử lại?"

├── [Thử lại] → Gửi lại yêu cầu

└── [Chơi offline] → Chế độ khách

```

  

#### Xung đột Đồng bộ Tiến trình

```

XUNG_DOT_DONG_BO

├── Tiến trình cục bộ > Tiến trình server

│ └── Tải lên cục bộ → Ghi đè server

│

└── Tiến trình server > Tiến trình cục bộ

└── Tải xuống server → Cập nhật cục bộ

```

  

### 10.2 Lỗi Tải Nội dung

  

```

TAI_BAI_HOC_THAT_BAI

├── Kiểm tra có phiên bản cache không

│ ├── [Có] → Tải cache

│ └── [Không] → Hiện lỗi

│ └── "Không thể tải bài học"

│ ├── [Thử lại] → Thử lại

│ └── [Quay lại] → Timeline

```

  

### 10.3 Trường hợp Đặc biệt với Tim

  

```

CAC_TRUONG_HOP_TIM

├── Khách hết tim, không có mạng

│ └── Không thể chơi cho đến khi có kết nối

│

├── Phát hiện thay đổi thời gian

│ └── Xác minh với thời gian server

│

└── Tim > tối đa (nghi ngờ hack)

└── Reset về tối đa (5)

```

  

---

## 11. KẾ HOẠCH NỘI DUNG RA MẮT

  

### 11.1 Nội dung Ban đầu (20 Đơn vị)

  

#### Nhân vật Huyền thoại (4)

1. **Hai Bà Trưng** - 5 bài học

2. **Ngô Quyền** - 4 bài học

3. **Trần Hưng Đạo** - 5 bài học

4. **Quang Trung** - 4 bài học

  

#### Nhân vật Hiếm (6)

1. **Đinh Bộ Lĩnh** - 3 bài học

2. **Lý Thường Kiệt** - 3 bài học

3. **Yết Kiêu** - 3 bài học

4. **Lê Lợi** - 3 bài học

5. **Nguyễn Trãi** - 2 bài học

6. **Gia Long** - 2 bài học

  

#### Sự kiện/Nhân vật Phổ thông (10)

1. **Trần Quốc Toản** - 1 bài học

2. **Lý Thái Tổ** - 1 bài học

3. **Trận Bạch Đằng 938** - 2 bài học

4. **Trận Chi Lăng** - 1 bài học

5. **Văn Miếu Quốc Tử Giám** - 1 bài học

6. **Hồ Quý Ly** - 1 bài học

7. **Mạc Đăng Dung** - 1 bài học

8. **Khởi nghĩa Tây Sơn** - 2 bài học

9. **Chiến tranh Trịnh-Nguyễn** - 1 bài học

10. **Phan Đình Phùng** - 1 bài học

  

**Tổng cộng**: ~50 bài học (150-250 phút nội dung)

  
  

*Hết Tài liệu*
