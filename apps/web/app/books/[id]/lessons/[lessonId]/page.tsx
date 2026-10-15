import Link from "next/link";
import { notFound } from "next/navigation";
import { ArrowLeft } from "lucide-react";
import {cookies} from "next/headers";
import ReaderActions from "./reader-actions";
type Lesson={id:string;title:string;content:string;content_type:string;book_id:string};

export default async function ReadLesson({params}:{params:Promise<{id:string;lessonId:string}>}){
 const {id,lessonId}=await params;
 const uuid=/^[a-fA-F0-9-]{36}$/;
 if(!uuid.test(id)||!uuid.test(lessonId))notFound();
 const base=process.env.API_INTERNAL_URL||"http://localhost:8080";
 const token=(await cookies()).get("openedu_session")?.value;
 const authHeaders=token?{Authorization:"Bearer "+token}:undefined;
 let lesson:Lesson;
 try{
  const response=await fetch(base+"/v1/lessons/"+encodeURIComponent(lessonId),{cache:"no-store",signal:AbortSignal.timeout(7000),headers:authHeaders});
  if(!response.ok)notFound();
  lesson=await response.json() as Lesson;
 }catch{notFound();}
 if(lesson.book_id!==id)notFound();
 let initial={completed:false,bookmarked:false};
 if(token){try{
  const res=await fetch(base+"/v1/books/"+encodeURIComponent(id)+"/progress",{cache:"no-store",signal:AbortSignal.timeout(7000),headers:authHeaders});
  if(res.ok){const data=await res.json() as {items:{lesson_id:string;completed:boolean;bookmarked:boolean}[]};
   const entry=data.items.find(x=>x.lesson_id===lessonId);if(entry)initial={completed:entry.completed,bookmarked:entry.bookmarked};}
 }catch{/* Reading remains available when personal progress is temporarily unavailable. */}}
 return <section className="container detail lesson-view">
  <Link href={"/books/"+encodeURIComponent(id)} className="back"><ArrowLeft size={16} aria-hidden="true"/>Danh mục chương và bài</Link>
  <article className="glass-card lesson-paper">
   <span className="eyebrow section-eyebrow">OPENEDU READER</span>
   <h1>{lesson.title}</h1>
   <div className="lesson-text">{lesson.content}</div>
   <ReaderActions lessonId={lessonId} initial={initial} signedIn={!!token}/>
   {lesson.content_type==="html"&&<p className="detail-notice">Nội dung HTML được hiển thị dưới dạng văn bản an toàn. Hình thức trình bày phong phú sẽ được hỗ trợ khi có quy trình lọc nội dung xuất bản.</p>}
  </article>
 </section>;
}