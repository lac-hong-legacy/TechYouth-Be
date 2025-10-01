package seeders

import (
	"log"
	"time"

	"github.com/lac-hong-legacy/ven_api/model"
	"gorm.io/gorm"
)

// TimelineSeeder handles seeding historical timelines
type TimelineSeeder struct {
	db *gorm.DB
}

// NewTimelineSeeder creates a new timeline seeder
func NewTimelineSeeder(db *gorm.DB) *TimelineSeeder {
	return &TimelineSeeder{db: db}
}

// SeedTimelines seeds the database with Vietnamese historical periods
func (s *TimelineSeeder) SeedTimelines() error {
	timelines := s.getHistoricalTimelines()

	for _, timeline := range timelines {
		// Check if timeline already exists
		var existingTimeline model.Timeline
		if err := s.db.Where("id = ?", timeline.ID).First(&existingTimeline).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Timeline doesn't exist, create it
				if err := s.db.Create(&timeline).Error; err != nil {
					log.Printf("Error creating timeline %s: %v", timeline.Era, err)
					return err
				}
				log.Printf("Created timeline: %s", timeline.Era)
			} else {
				log.Printf("Error checking timeline %s: %v", timeline.Era, err)
				return err
			}
		} else {
			log.Printf("Timeline %s already exists, skipping", timeline.Era)
		}
	}

	log.Println("Timeline seeding completed successfully")
	return nil
}

// getHistoricalTimelines returns the Vietnamese historical periods
func (s *TimelineSeeder) getHistoricalTimelines() []model.Timeline {
	now := time.Now()

	timelines := []model.Timeline{
		{
			ID:          "timeline_van_lang",
			Era:         "Bac_Thuoc",
			Dynasty:     "Văn Lang",
			StartYear:   -2879,
			Order:       1,
			EndYear:     intPtr(-258),
			Description: "Vương quốc Việt Nam huyền thoại đầu tiên do Hùng Vương sáng lập. Thời kỳ nền văn minh Lạc Việt và văn hóa thời đại đồ đồng.",
			KeyEvents: jsonArray([]string{
				"Sáng lập vương quốc Văn Lang",
				"Thiết lập triều đại Hùng Vương",
				"Phát triển văn hóa thời đại đồ đồng",
				"Tiến bộ trong trồng lúa",
			}),
			CharacterIds: jsonArray([]string{"char_hung_vuong"}),
			ImageURL:     "/assets/timeline/van_lang.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_au_lac",
			Era:         "Bac_Thuoc",
			Dynasty:     "Âu Lạc",
			StartYear:   -257,
			Order:       2,
			EndYear:     intPtr(-207),
			Description: "Vương quốc do An Dương Vương thành lập sau khi chinh phục Văn Lang. Nổi tiếng với việc xây dựng thành Cổ Loa.",
			KeyEvents: jsonArray([]string{
				"An Dương Vương chinh phục Văn Lang",
				"Thành lập vương quốc Âu Lạc",
				"Xây dựng thành Cổ Loa",
				"Đưa vào công nghệ nỏ",
			}),
			CharacterIds: jsonArray([]string{}),
			ImageURL:     "/assets/timeline/au_lac.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_bac_thuoc",
			Era:         "Bac_Thuoc",
			Dynasty:     "Bắc thuộc",
			StartYear:   -111,
			Order:       3,
			EndYear:     intPtr(938),
			Description: "Thời kỳ Trung Quốc thống trị kéo dài hơn 1000 năm, bị gián đoạn bởi một số cuộc nổi dậy lớn bao gồm Hai Bà Trưng và Bà Triệu.",
			KeyEvents: jsonArray([]string{
				"Cuộc chinh phục của nhà Hán Trung Quốc (-111)",
				"Cuộc nổi dậy của Hai Bà Trưng (40-43 SCN)",
				"Cuộc nổi dậy của Bà Triệu (248 SCN)",
				"Các phong trào kháng chiến khác nhau",
				"Giao lưu văn hóa và công nghệ",
			}),
			CharacterIds: jsonArray([]string{"char_hai_ba_trung", "char_ba_trieu"}),
			ImageURL:     "/assets/timeline/bac_thuoc.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_ngo",
			Era:         "Doc_Lap",
			Dynasty:     "Ngô",
			StartYear:   939,
			Order:       4,
			EndYear:     intPtr(965),
			Description: "Triều đại Việt Nam độc lập đầu tiên sau sự cai trị của Trung Quốc, do Ngô Quyền sáng lập sau chiến thắng tại sông Bạch Đằng.",
			KeyEvents: jsonArray([]string{
				"Trận Bạch Đằng (938)",
				"Chấm dứt sự thống trị của Trung Quốc",
				"Thành lập triều đại Ngô",
				"Kinh đô tại Cổ Loa",
			}),
			CharacterIds: jsonArray([]string{"char_ngo_quyen"}),
			ImageURL:     "/assets/timeline/ngo.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_dinh_le",
			Era:         "Doc_Lap",
			Dynasty:     "Đinh - Tiền Lê",
			StartYear:   968,
			Order:       5,
			EndYear:     intPtr(1009),
			Description: "Thời kỳ của các triều đại Đinh và Tiền Lê, thiết lập nền tảng cho cấu trúc nhà nước Việt Nam độc lập.",
			KeyEvents: jsonArray([]string{
				"Đinh Bộ Lĩnh thống nhất Việt Nam",
				"Thành lập Đại Cồ Việt",
				"Dời đô về Hoa Lư",
				"Triều đại Tiền Lê",
			}),
			CharacterIds: jsonArray([]string{"char_dinh_bo_linh"}),
			ImageURL:     "/assets/timeline/dinh_le.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_ly",
			Era:         "Phong_Kien",
			Dynasty:     "Lý",
			StartYear:   1009,
			Order:       6,
			EndYear:     intPtr(1225),
			Description: "Thời kỳ hoàng kim của văn hóa Việt Nam và Phật giáo. Dời đô về Thăng Long (Hà Nội). Thời kỳ mở rộng lãnh thổ và phát triển văn hóa.",
			KeyEvents: jsonArray([]string{
				"Lý Thái Tổ sáng lập triều đại (1009)",
				"Dời đô về Thăng Long (1010)",
				"Thành lập Văn Miếu",
				"Chiến thắng nhà Tống Trung Quốc",
				"Phật giáo văn hóa thịnh vượng",
			}),
			CharacterIds: jsonArray([]string{"char_ly_thai_to", "char_ly_thuong_kiet"}),
			ImageURL:     "/assets/timeline/ly.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_tran",
			Era:         "Phong_Kien",
			Dynasty:     "Trần",
			StartYear:   1225,
			Order:       7,
			EndYear:     intPtr(1400),
			Description: "Triều đại nổi tiếng với việc đẩy lùi thành công ba cuộc xâm lược của quân Mông Cổ dưới sự lãnh đạo của Trần Hưng Đạo. Thời kỳ đổi mới quân sự và phát triển văn hóa.",
			KeyEvents: jsonArray([]string{
				"Triều đại Trần được thành lập (1225)",
				"Đẩy lùi cuộc xâm lược Mông Cổ lần thứ nhất (1258)",
				"Đánh bại cuộc xâm lược Mông Cổ lần thứ hai (1285)",
				"Nghiền nát cuộc xâm lược Mông Cổ lần thứ ba (1287-1288)",
				"Phát triển chiến thuật du kích",
			}),
			CharacterIds: jsonArray([]string{"char_tran_hung_dao"}),
			ImageURL:     "/assets/timeline/tran.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_ho",
			Era:         "Phong_Kien",
			Dynasty:     "Hồ",
			StartYear:   1400,
			Order:       8,
			EndYear:     intPtr(1407),
			Description: "Triều đại tồn tại ngắn ngủi nổi tiếng với các cải cách tiến bộ nhưng kết thúc với cuộc xâm lược của nhà Minh Trung Quốc do xung đột nội bộ.",
			KeyEvents: jsonArray([]string{
				"Hồ Quý Ly chiếm quyền (1400)",
				"Cải cách đất đai và tiền tệ",
				"Các cuộc nổi dậy nội bộ",
				"Cuộc xâm lược của nhà Minh Trung Quốc (1407)",
			}),
			CharacterIds: jsonArray([]string{"char_ho_quy_ly"}),
			ImageURL:     "/assets/timeline/ho.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_ming_occupation",
			Era:         "Phong_Kien",
			Dynasty:     "Minh chiếm đóng",
			StartYear:   1407,
			Order:       9,
			EndYear:     intPtr(1428),
			Description: "Thời kỳ nhà Minh Trung Quốc chiếm đóng. Sự kháng chiến của người Việt do Lê Lợi lãnh đạo đạt đỉnh điểm với việc giành độc lập và sáng lập triều đại Lê.",
			KeyEvents: jsonArray([]string{
				"Sự chiếm đóng của nhà Minh Trung Quốc bắt đầu (1407)",
				"Chính sách đàn áp văn hóa",
				"Lê Lợi bắt đầu kháng chiến (1418)",
				"Khởi nghĩa Lam Sơn",
				"Giải phóng Việt Nam (1428)",
			}),
			CharacterIds: jsonArray([]string{"char_le_loi", "char_nguyen_trai"}),
			ImageURL:     "/assets/timeline/ming_occupation.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_later_le",
			Era:         "Phong_Kien",
			Dynasty:     "Hậu Lê",
			StartYear:   1428,
			Order:       10,
			EndYear:     intPtr(1788),
			Description: "Triều đại dài nhất trong lịch sử Việt Nam. Thời kỳ hoàng kim dưới Lê Thánh Tông với việc mở rộng lãnh thổ và hệ thống hóa pháp luật.",
			KeyEvents: jsonArray([]string{
				"Triều đại Lê được thành lập (1428)",
				"Thời kỳ hoàng kim Hồng Đức (1470-1497)",
				"Chinh phục các lãnh thổ Chăm",
				"Bộ luật Hồng Đức",
				"Phân chia thành thời kỳ Trịnh-Nguyễn",
			}),
			CharacterIds: jsonArray([]string{"char_le_thanh_tong"}),
			ImageURL:     "/assets/timeline/later_le.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_mac",
			Era:         "Phong_Kien",
			Dynasty:     "Mạc",
			StartYear:   1527,
			Order:       11,
			EndYear:     intPtr(1677),
			Description: "Triều đại kiểm soát miền Bắc Việt Nam trong thời kỳ phân chia, cạnh tranh với triều đại Lê phục hồi.",
			KeyEvents: jsonArray([]string{
				"Mạc Đăng Dung chiếm quyền (1527)",
				"Kiểm soát các lãnh thổ phía bắc",
				"Xung đột với triều đại Lê phục hồi",
				"Mất dần lãnh thổ",
			}),
			CharacterIds: jsonArray([]string{"char_mac_dang_dung"}),
			ImageURL:     "/assets/timeline/mac.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_tay_son",
			Era:         "Phong_Kien",
			Dynasty:     "Tây Sơn",
			StartYear:   1778,
			Order:       12,
			EndYear:     intPtr(1802),
			Description: "Thời kỳ cách mạng do anh em họ Nguyễn lãnh đạo. Quang Trung đánh bại cuộc xâm lược của nhà Thanh Trung Quốc và thực hiện các cải cách tiến bộ.",
			KeyEvents: jsonArray([]string{
				"Khởi nghĩa Tây Sơn bắt đầu (1778)",
				"Đánh bại các chúa Trịnh và Nguyễn",
				"Quang Trung trở thành hoàng đế (1788)",
				"Chiến thắng cuộc xâm lược của nhà Thanh (1789)",
				"Cải cách xã hội và giáo dục",
			}),
			CharacterIds: jsonArray([]string{"char_nguyen_hue"}),
			ImageURL:     "/assets/timeline/tay_son.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_nguyen",
			Era:         "Phong_Kien",
			Dynasty:     "Nguyễn",
			StartYear:   1802,
			Order:       13,
			EndYear:     intPtr(1945),
			Description: "Triều đại hoàng gia cuối cùng của Việt Nam. Thời kỳ thống nhất, những nỗ lực hiện đại hóa và cuối cùng là thực dân Pháp.",
			KeyEvents: jsonArray([]string{
				"Gia Long thống nhất Việt Nam (1802)",
				"Kinh đô được thành lập tại Huế",
				"Thời kỳ thực dân Pháp bắt đầu (1858)",
				"Các phong trào kháng chiến Cần Vương",
				"Kết thúc chế độ hoàng gia (1945)",
			}),
			CharacterIds: jsonArray([]string{"char_nguyen_anh", "char_nguyen_du"}),
			ImageURL:     "/assets/timeline/nguyen.jpg",
			CreatedAt:    now,
			UpdatedAt:    now,
		},
		{
			ID:          "timeline_modern",
			Era:         "Can_Dai",
			Dynasty:     "Cận đại",
			StartYear:   1858,
			Order:       14,
			EndYear:     intPtr(1975),
			Description: "Thời kỳ thực dân Pháp, các phong trào độc lập, và các cuộc chiến tranh giải phóng kết thúc bằng việc thống nhất đất nước.",
			KeyEvents: jsonArray([]string{
				"Cuộc chinh phục của Pháp bắt đầu (1858)",
				"Các phong trào dân tộc chủ nghĩa nổi lên",
				"Thế chiến II và sự chiếm đóng của Nhật Bản",
				"Tuyên bố Độc lập (1945)",
				"Chiến tranh Đông Dương lần thứ nhất (1946-1954)",
				"Chiến tranh Việt Nam (1955-1975)",
				"Thống nhất Việt Nam (1975)",
			}),
			CharacterIds: jsonArray([]string{
				"char_phan_boi_chau",
				"char_phan_chu_trinh",
				"char_vo_thi_sau",
				"char_ho_chi_minh",
			}),
			ImageURL:  "/assets/timeline/modern.jpg",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	return timelines
}
