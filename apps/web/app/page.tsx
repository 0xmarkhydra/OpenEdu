import Link from "next/link";

type Book = {
  id: string;
  title: string;
  grade: number;
  subject_code: string;
  rights_owner: string;
  rights_license: string;
};

async function loadBooks(query: URLSearchParams): Promise<{ items: Book[]; offline: boolean }> {
  const url = new URL("/v1/books", process.env.API_INTERNAL_URL || "http://localhost:8080");
  url.search = query.toString();
  try {
    const response = await fetch(url.toString(), { cache: "no-store", signal: AbortSignal.timeout(5000) });
    if (!response.ok) return { items: [], offline: true };
    const json: unknown = await response.json();
    if (typeof json !== "object" || !json || !("items" in json) || !Array.isArray(json.items)) {
      return { items: [], offline: true };
    }
    return { items: json.items as Book[], offline: false };
  } catch {
    return { items: [], offline: true };
  }
}

export default async function LibraryPage({
  searchParams,
}: {
  searchParams: Promise<{ q?: string; grade?: string; subject?: string }>;
}) {
  const params = await searchParams;
  const q = typeof params.q === "string" ? params.q.slice(0, 120) : "";
  const grade = typeof params.grade === "string" && /^(1[0-2]|[1-9])$/.test(params.grade) ? params.grade : "";
  const subject = typeof params.subject === "string" ? params.subject.slice(0, 80) : "";
  const query = new URLSearchParams();
  if (q) query.set("q", q);
  if (grade) query.set("grade", grade);
  if (subject) query.set("subject", subject);
  const { items, offline } = await loadBooks(query);

  return (
    <>
      <section className="hero">
        <div className="container hero-inner">
          <span className="eyebrow">FREE EDUCATION CORE</span>
          <h1>Kho tri thức mở ra<br /><span>mỗi ngày học mới.</span></h1>
          <p>Khám phá sách và bài học được phép phân phối — tìm đúng nội dung phù hợp với lớp và môn học của bạn.</p>
          <form className="search-panel" action="/" method="get" role="search">
            <label htmlFor="q" className="sr-only">Tìm sách giáo khoa, bài học</label>
            <input name="q" id="q" type="search" placeholder="Tìm sách, bài học, từ khóa..." defaultValue={q} maxLength={120} />
            <label htmlFor="grade" className="sr-only">Chọn lớp</label>
            <select name="grade" id="grade" defaultValue={grade} aria-label="Khối lớp">
              <option value="">Tất cả lớp</option>
              {Array.from({ length: 12 }, (_, i) => <option value={i + 1} key={i}>Lớp {i + 1}</option>)}
            </select>
            <button type="submit">Tìm học liệu →</button>
          </form>
          <div className="hero-caption">Sách và học liệu chỉ được công bố sau khi xác thực quyền sử dụng.</div>
        </div>
      </section>

      <section className="container library" aria-labelledby="library-title">
        <div className="section-heading">
          <div><span className="eyebrow section-eyebrow">THƯ VIỆN SỐ</span><h2 id="library-title">Học liệu đang phát hành</h2></div>
          <span className="count">{items.length} tài liệu</span>
        </div>
        {offline ? (
          <div className="empty-state" role="status"><h3>Chưa thể kết nối hệ thống học liệu</h3><p>Máy chủ API hoặc cơ sở dữ liệu chưa sẵn sàng. Hãy thử lại sau.</p></div>
        ) : items.length === 0 ? (
          <div className="empty-state"><h3>Chưa có học liệu phù hợp</h3><p>Thử tìm theo tên sách hoặc chọn một khối lớp khác. Tài liệu chỉ hiển thị khi đã được phê duyệt và phát hành.</p></div>
        ) : (
          <div className="book-grid">
            {items.map(book => (
              <Link href={"/books/" + encodeURIComponent(book.id)} className="book-card" key={book.id}>
                <div className="book-cover" aria-hidden="true"><span className="cover-icon">✦</span><span>LỚP {book.grade}</span></div>
                <div className="book-meta"><span className="subject-chip">{book.subject_code}</span><h3>{book.title}</h3><p>{book.rights_owner}</p><span className="view-label">Xem học liệu ↗</span></div>
              </Link>
            ))}
          </div>
        )}
      </section>
      <section className="container features">
        <div><span className="feature-icon">📚</span><h3>Học liệu chuẩn hóa</h3><p>Sắp xếp theo môn học, lớp và bài để tìm kiếm thuận tiện.</p></div>
        <div><span className="feature-icon">📝</span><h3>Học và đánh giá</h3><p>Hệ thống bài tập, kết quả và tiến độ đang được phát triển.</p></div>
        <div><span className="feature-icon">🤝</span><h3>Kết nối gia đình</h3><p>Tương lai hỗ trợ nhà trường, giáo viên và phụ huynh.</p></div>
      </section>
    </>
  );
}
