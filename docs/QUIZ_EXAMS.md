# OpenEdu v0.6 — Bài kiểm tra trắc nghiệm có chấm điểm tự động

## Giới hạn đã triển khai

- Giáo viên **được phân công lớp** hoặc ban giám hiệu/quản trị trường tạo đề trắc nghiệm từ 5–180 phút; ngày kết thúc bắt buộc phải nằm trong tương lai.
- Thêm tối đa 100 câu hỏi cho mỗi đề; mỗi câu gồm bốn đáp án và một chỉ số đáp án chính xác. Câu hỏi chỉ thêm ở bản nháp.
- Phải có tối thiểu một câu hỏi mới được phát hành. Sau khi phát hành, không có API sửa đáp án hay thay câu hỏi để tránh gian lận/chênh lệch giữa các lượt thi.
- Học sinh **đang ghi danh lớp** chỉ nhìn thấy đề đã phát hành. Muốn xem nội dung phải `POST .../start`; PostgreSQL đặt giờ hết hạn tối đa là thời lượng đề và không sau hạn cuối.
- Mỗi học sinh chỉ có một bản ghi lượt làm cho một đề. Bấm `start` nhiều lần không gia hạn đồng hồ. Nộp bài một lần, phải trả lời toàn bộ câu, máy chủ chấm thang điểm 10; mỗi câu có cùng trọng số.
- Không có endpoint đưa đáp án đúng cho học sinh, kể cả sau khi hoàn thành; chỉ giáo viên được phân công/lãnh đạo trường có thể xem khóa đáp án và kết quả tổng hợp.
- Nộp sau khi hết giờ, nộp trùng, nộp không đủ câu hay giả mạo ID câu hỏi đều bị từ chối. Kết quả thí sinh không bao gồm đáp án đúng. Điểm được lưu theo từng học sinh/đề.
- Phụ huynh không có quyền tham gia bài kiểm tra thay học sinh. Sau khi được xác minh quan hệ, phụ huynh chỉ xem **điểm thi đã nộp của con mình** tại Sổ liên lạc; không xem đáp án/câu trả lời hoặc điểm của học sinh khác. Thu hồi quan hệ giám hộ sẽ thu hồi luôn quyền xem kết quả.

## API

| Method | URL | Quyền |
| --- | --- | --- |
| GET | /v1/classes/{id}/quizzes | Giáo viên lớp, lãnh đạo trường hoặc học sinh đang ghi danh |
| POST | /v1/classes/{id}/quizzes | Giáo viên lớp hoặc lãnh đạo trường; `{"title":"Kiểm tra 15 phút","duration_minutes":15,"due_at":"2026-11-01T00:00:00Z"}` |
| POST | /v1/quizzes/{id}/questions | Giáo viên lớp hoặc lãnh đạo trường; `{"position":1,"prompt":"2 + 2 = ?","options":["2","3","4","5"],"correct_option":2}` |
| POST | /v1/quizzes/{id}/publish | Giáo viên lớp hoặc lãnh đạo trường |
| GET | /v1/quizzes/{id} | Giáo viên thấy khóa đáp án; học sinh chỉ thấy câu hỏi trong thời gian lượt làm còn hiệu lực |
| POST | /v1/quizzes/{id}/start | Học sinh đang ghi danh lớp; không tạo lượt mới khi đã có lượt |
| POST | /v1/quizzes/{id}/submit | Học sinh; `{"answers":{"UUID-QUESTION-1":2,"UUID-QUESTION-2":1}}` |
| GET | /v1/quizzes/{id}/results | Giáo viên lớp hoặc lãnh đạo trường |

Web UI: vào **Dashboard → Kiểm tra**. Native Flutter: vào tab **Lớp học**, mở lớp rồi chọn bài kiểm tra. Flutter hỗ trợ học sinh làm bài/chấm tự động; soạn đề và xem kết quả giáo viên vẫn dùng web. Cả hai tuân thủ icon Lucide, iOS Liquid Glass.

## Kiểm thử

- Disposable PostgreSQL integration checks teacher/learner role boundaries, draft secrecy, no answer leak, correct/incorrect answers, 5/10 grading, immutable published questions, duplicate attempts and submission refusal after expiry.
- Go unit tests and vet. Next.js compiler/TypeScript and icon scan before push.
- Không dùng dữ liệu học sinh thật, tài khoản mặc định, tài liệu/bài tập vi phạm bản quyền trong test.

## Chưa thuộc phạm vi hoàn thành

Ngân hàng câu hỏi tái sử dụng giữa đề, câu tự luận, tệp đính kèm, thang đo riêng, random câu hỏi, giám sát thi, phát hiện gian lận, quyền thi lại/đặc cách, xem lại đáp án và trích xuất báo cáo. Cần chính sách bảo mật/đồng ý của nhà trường trước khi xử lý dữ liệu trẻ em thực.
