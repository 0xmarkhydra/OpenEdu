"use client";
import {useState} from "react";
import Link from "next/link";
import {Bookmark,BookmarkCheck,CheckCheck,CheckCircle2,LoaderCircle,LogIn} from "lucide-react";
import {Button} from "@/components/ui/button";

type ReadingState={completed:boolean;bookmarked:boolean};
export default function ReaderActions({lessonId,initial,signedIn}:{
 lessonId:string;initial:ReadingState;signedIn:boolean
}){
 const [state,setState]=useState(initial);
 const [pending,setPending]=useState(false);
 const [error,setError]=useState("");
 async function save(changes:Partial<ReadingState>){
  if(pending)return;
  const updated={...state,...changes};
  setPending(true);setError("");
  try{
   const r=await fetch("/api/reading/lessons/"+encodeURIComponent(lessonId)+"/progress",{
    method:"PUT",headers:{"Content-Type":"application/json"},body:JSON.stringify(updated)
   });
   const payload=await r.json();
   if(!r.ok)throw new Error(payload.error||"Không thể lưu tiến độ");
   setState({completed:!!payload.completed,bookmarked:!!payload.bookmarked});
  }catch(e){setError(e instanceof Error?e.message:"Không thể lưu tiến độ");}
  finally{setPending(false);}
 }
 if(!signedIn)return <div className="reader-actions"><p>Đăng nhập để lưu dấu trang và tiến độ học tập của riêng bạn.</p>
  <Button asChild variant="outline"><Link href="/login"><LogIn size={17} aria-hidden="true"/>Đăng nhập</Link></Button></div>;
 return <div className="reader-actions" aria-label="Trạng thái học tập">
  <div className="reader-action-buttons">
   <Button disabled={pending} variant={state.bookmarked?"default":"outline"} onClick={()=>void save({bookmarked:!state.bookmarked})}>
    {state.bookmarked?<BookmarkCheck size={17} aria-hidden="true"/>:<Bookmark size={17} aria-hidden="true"/>}
    {state.bookmarked?"Đã lưu bài học":"Lưu bài học"}
   </Button>
   <Button disabled={pending} variant={state.completed?"secondary":"default"} onClick={()=>void save({completed:!state.completed})}>
    {pending?<LoaderCircle size={17} aria-hidden="true"/>:state.completed?<CheckCheck size={17} aria-hidden="true"/>:<CheckCircle2 size={17} aria-hidden="true"/>}
    {state.completed?"Đã học xong":"Đánh dấu đã học"}
   </Button>
  </div>
  {error&&<p role="alert" className="portal-message">{error}</p>}
  {state.completed&&<p className="reader-status" role="status"><CheckCheck size={17} aria-hidden="true"/>Bài học này đã được ghi nhận hoàn thành.</p>}
 </div>;
}
