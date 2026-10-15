import Link from "next/link";
import { notFound } from "next/navigation";

type Book={id:string;title:string;grade:number;subject_code:string;rights_owner:string;rights_license:string;rights_source:string;status:string};
type Lesson={id:string;title:string;position:number;content_type:string};
type Chapter={id:string;title:string;position:number;lessons:Lesson[]};

export default async function BookPage({params}:{params:Promise<{id:string}>}){
 const {id}=await params;
 if(!/^[a-fA-F0-9-]{36}$/.test(id))notFound();
 const endpoint=process.env.API_INTERNAL_URL||"http://localhost:8080";
 const [bookResponse,outlineResponse]=await Promise.all([
  fetch(endpoint+"/v1/books/"+encodeURIComponent(id),{cache:"no-store",signal:AbortSignal.timeout(7000)}).catch(()=>null),
  fetch(endpoint+"/v1/books/"+encodeURIComponent(id)+"/outline",{cache:"no-store",signal:AbortSignal.timeout(7000)}).catch(()=>null)
 ]);
 if(!bookResponse?.ok)notFound();
 const book=await bookResponse.json() as Book;
 const chapters:Chapter[]=outlineResponse?.ok?(await outlineResponse.json() as {items:Chapter[]}).items||[]:[];
 return <section className="container detail">
  <Link href="/" className="back">← Thư viện OpenEdu</Link>
  <div className="detail-card">
   <div className="detail-cover" aria-hidden="true">📖</div>
   <div>
    <span className="eyebrow section-eyebrow">HỌC LIỆU ĐƯỢC PHÁT HÀNH · LỚP {book.grade}</span>
    <h1>{book.title}</h1><p className="detail-subject">Môn học: {book.subject_code}</p>
    <dl><dt>Đơn vị sở hữu</dt><dd>{book.rights_owner}</dd>
     <dt>Giấy phép sử dụng</dt><dd>{book.rights_license}</dd>
     <dt>Nguồn học liệu</dt><dd>{book.rights_source}</dd></dl>
    <p className="detail-notice">Nội dung học liệu được hiển thị theo phạm vi quyền phân phối. PDF có bản quyền sẽ chỉ được mở khi có bộ lưu trữ và quyền truy cập hợp lệ.</p>
   </div>
  </div>
  <div className="glass-card book-outline">
   <span className="eyebrow section-eyebrow">NỘI DUNG CHƯƠNG TRÌNH</span><h2>Danh mục bài học</h2>
   {chapters.length===0?<div className="empty-state"><h3>Chưa có bài học</h3><p>Nhà trường hoặc tác giả chưa phát hành nội dung chương/bài cho sách này.</p></div>:chapters.map(ch=>
    <section className="chapter-section" key={ch.id}>
     <h3><span>{String(ch.position).padStart(2,"0")}</span> {ch.title}</h3>
     {ch.lessons.length===0?<p className="muted">Chưa có bài học trong chương.</p>:<div className="outline-lessons">{ch.lessons.map(lesson=>
      lesson.content_type==="pdf"
      ? <span className="outline-link disabled" key={lesson.id}><span>{lesson.position}. {lesson.title}</span><small>PDF — chờ cấp quyền</small></span>
      : <Link className="outline-link" key={lesson.id} href={"/books/"+encodeURIComponent(id)+"/lessons/"+encodeURIComponent(lesson.id)}><span>{lesson.position}. {lesson.title}</span><span aria-hidden="true">›</span></Link>
     )}</div>}
    </section>
   )}
  </div>
 </section>;
}
