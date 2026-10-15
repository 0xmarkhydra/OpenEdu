"use client";
import { FormEvent,useCallback,useEffect,useState } from "react";
type Book={id:string;title:string;grade:number;subject_code:string;status:string;visibility:string};
type Lesson={id:string;title:string;position:number};
type Chapter={id:string;title:string;position:number;lessons:Lesson[]};
type Item<T>={items:T[]};

export default function ContentEditor({global}:{global:boolean}){
 const [books,setBooks]=useState<Book[]>([]);
 const [selected,setSelected]=useState("");
 const [chapters,setChapters]=useState<Chapter[]>([]);
 const [chapterId,setChapterId]=useState("");
 const [title,setTitle]=useState("");
 const [grade,setGrade]=useState(6);
 const [subject,setSubject]=useState("toan");
 const [owner,setOwner]=useState("");
 const [license,setLicense]=useState("");
 const [source,setSource]=useState("");
 const [chapterTitle,setChapterTitle]=useState("");
 const [chapterPosition,setChapterPosition]=useState(1);
 const [lessonTitle,setLessonTitle]=useState("");
 const [lessonPosition,setLessonPosition]=useState(1);
 const [content,setContent]=useState("");
 const [message,setMessage]=useState("");
 const [busy,setBusy]=useState(false);
 const canEdit=(book:Book|undefined)=>book?.status==="draft";
 const current=books.find(b=>b.id===selected);
 const api=useCallback(async<T,>(path:string,method="GET",body?:object):Promise<T>=>{
  const res=await fetch("/api/catalog/"+path,{method,cache:"no-store",headers:{"content-type":"application/json"},body:body?JSON.stringify(body):undefined});
  const data=await res.json() as T & {error?:string};
  if(!res.ok)throw new Error(data.error||"Không thể xử lý học liệu");
  return data;
 },[]);
 const refresh=useCallback(async()=>{
  try{const value=await api<Item<Book>>("books");setBooks(value.items||[])}
  catch(err){setMessage(err instanceof Error?err.message:"Không thể tải sách");}
 },[api]);
 const refreshOutline=useCallback(async(id:string)=>{
  const selectedBook=books.find(b=>b.id===id);
  if(!id||selectedBook?.status!=="draft"){setChapters([]);return;}
  try{const value=await api<Item<Chapter>>("books/"+encodeURIComponent(id)+"/chapters");setChapters(value.items||[])}
  catch(err){setMessage(err instanceof Error?err.message:"Không thể tải chương sách");}
 },[api,books]);
 useEffect(()=>{void refresh()},[refresh]);
 useEffect(()=>{void refreshOutline(selected)},[selected,refreshOutline]);
 const run=async(path:string,body:object,success:string,after?:()=>void)=>{
  setBusy(true);setMessage("");
  try{
   const result=await api<{id?:string}>(path,"POST",body);
   setMessage(success);after?.();await refresh();
   if(selected)await refreshOutline(selected);
   return result;
  }catch(err){setMessage(err instanceof Error?err.message:"Không thể lưu");return null}
  finally{setBusy(false)}
 };
 async function create(event:FormEvent<HTMLFormElement>){
  event.preventDefault();
  const value=await run("books",{title,grade,subject_code:subject,rights_owner:owner,rights_license:license,rights_source:source,visibility:global?"public":"school"},"Đã tạo bản nháp. Thêm chương và bài trước khi gửi duyệt.",()=>{setTitle("");});
  if(value?.id)setSelected(value.id);
 }
 const post=(event:FormEvent<HTMLFormElement>,path:string,body:object,notice:string,after?:()=>void)=>{
  event.preventDefault();void run(path,body,notice,after)
 };
 return <div className="portal-two content-studio">
  {message&&<div className="portal-message glass-wide" role="status">{message}</div>}
  <section className="glass-card">
   <h2>Danh sách học liệu</h2><p className="muted">Bản nháp → Gửi duyệt → Duyệt → Phát hành. Không chỉnh sửa trực tiếp bản đã công bố.</p>
   {books.map(b=><button className={"class-item "+(selected===b.id?"selected":"")} key={b.id} onClick={()=>{setSelected(b.id);setChapterId("");}}>
    {b.title}<small>Lớp {b.grade} · {b.subject_code} · {b.status}</small>
   </button>)}
   {!books.length&&<p className="muted">Chưa có bản nháp học liệu.</p>}
   <form className="portal-form" onSubmit={create}>
    <h3>Tạo học liệu mới</h3>
    <label>Tên sách / tài liệu<input required maxLength={160} value={title} onChange={e=>setTitle(e.target.value)}/></label>
    <label>Khối lớp<select value={grade} onChange={e=>setGrade(Number(e.target.value))}>{Array.from({length:12},(_,i)=><option value={i+1} key={i}>Lớp {i+1}</option>)}</select></label>
    <label>Mã môn học<input required maxLength={60} value={subject} onChange={e=>setSubject(e.target.value)}/></label>
    <label>Chủ sở hữu quyền<input required maxLength={500} value={owner} onChange={e=>setOwner(e.target.value)}/></label>
    <label>Giấy phép phân phối<input required maxLength={500} value={license} onChange={e=>setLicense(e.target.value)}/></label>
    <label>Nguồn cấp quyền<input required maxLength={500} value={source} onChange={e=>setSource(e.target.value)}/></label>
    <p className="muted">{global?"Học liệu công khai cần quyền xuất bản toàn hệ thống.":"Học liệu chỉ hiển thị trong phạm vi trường đã cấp quyền."} Không đăng nội dung có bản quyền nếu chưa có phép.</p>
    <button disabled={busy}>Tạo bản nháp</button>
   </form>
  </section>
  <section className="glass-card">
   <h2>Nội dung và xuất bản</h2>
   {!current?<p className="muted">Chọn một học liệu để soạn chương và bài học.</p>:<>
    <div className="portal-row"><strong>{current.title}</strong><span className="pill">{current.status}</span></div>
    {current.status==="draft"&&<>
     <div className="chapter-section">
      <h3>Danh sách chương và bài</h3>
      {chapters.map(ch=><div className="portal-row" key={ch.id}><div><strong>{ch.position}. {ch.title}</strong><small>{ch.lessons.length} bài học</small></div><button className="outline-action" onClick={()=>setChapterId(ch.id)}>{chapterId===ch.id?"Đã chọn":"Chọn"}</button></div>)}
      {!chapters.length&&<p className="muted">Chưa có chương.</p>}
     </div>
     <form className="portal-form" onSubmit={e=>post(e,"books/"+encodeURIComponent(current.id)+"/chapters",{title:chapterTitle,position:chapterPosition},"Đã tạo chương",()=>{setChapterTitle("");setChapterPosition(v=>v+1);})}>
      <h3>Thêm chương</h3><label>Tên chương<input value={chapterTitle} required onChange={e=>setChapterTitle(e.target.value)}/></label>
      <label>Vị trí<input type="number" min="1" max="500" value={chapterPosition} onChange={e=>setChapterPosition(Number(e.target.value))}/></label>
      <button disabled={busy}>Thêm chương</button>
     </form>
     <form className="portal-form" onSubmit={e=>post(e,"chapters/"+encodeURIComponent(chapterId)+"/lessons",{title:lessonTitle,position:lessonPosition,content},"Đã thêm bài học",()=>{setLessonTitle("");setContent("");setLessonPosition(v=>v+1);})}>
      <h3>Thêm bài học</h3><label>Chương<select required value={chapterId} onChange={e=>setChapterId(e.target.value)}><option value="">Chọn chương</option>{chapters.map(ch=><option key={ch.id} value={ch.id}>{ch.title}</option>)}</select></label>
      <label>Tiêu đề bài<input required maxLength={180} value={lessonTitle} onChange={e=>setLessonTitle(e.target.value)}/></label>
      <label>Vị trí bài<input type="number" min="1" max="500" value={lessonPosition} onChange={e=>setLessonPosition(Number(e.target.value))}/></label>
      <label>Nội dung có quyền phân phối<textarea required rows={8} value={content} onChange={e=>setContent(e.target.value)}/></label>
      <button disabled={busy||!chapterId}>Lưu bài học</button>
     </form>
    </>}
    <div className="portal-form">
     <h3>Quy trình kiểm duyệt</h3>
     <p className="muted">Hệ thống chỉ cho người có vai trò phù hợp chuyển trạng thái. Học liệu được bảo vệ theo giấy phép và trường học.</p>
     {current.status==="draft"&&<button onClick={()=>void run("books/"+current.id+"/submit",{},"Đã chuyển sang chờ duyệt")} disabled={busy}>Gửi duyệt</button>}
     {current.status==="review"&&<button onClick={()=>void run("books/"+current.id+"/approve",{},"Đã phê duyệt học liệu")} disabled={busy}>Phê duyệt</button>}
     {current.status==="approved"&&<button onClick={()=>void run("books/"+current.id+"/publish",{},"Đã xuất bản học liệu")} disabled={busy}>Xuất bản</button>}
     {current.status==="published"&&<button onClick={()=>void run("books/"+current.id+"/archive",{},"Đã lưu trữ học liệu")} disabled={busy}>Ngừng phát hành</button>}
     {current.status==="archived"&&<p className="muted">Bản này đã kết thúc phát hành.</p>}
    </div>
   </>}
  </section>
 </div>;
}
