"use client";

import {FormEvent,useCallback,useEffect,useState} from "react";
import {AlarmClock,CheckCircle2,ClipboardList,FileQuestion,Plus,Send,ShieldCheck,UsersRound} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Card} from "@/components/ui/card";

type SchoolClass={id:string;name:string;grade:number};
type Quiz={id:string;title:string;status:string;duration_minutes:number;due_at:string;questions:number;grade:number|null;submitted:boolean};
type Question={id:string;position:number;prompt:string;options:string[];correct_option?:number};
type QuizDetail={id:string;status:string;due_at:string;started:boolean;expires_at:string|null;submitted:boolean;grade:number|null;items:Question[]};
type Result={student_id:string;name:string;started_at:string;submitted:boolean;grade:number|null};

export default function QuizPanel({classes,roles}:{classes:SchoolClass[];roles:string[]}){
 const staff=roles.some(r=>["teacher","school_admin","principal"].includes(r));
 const [classId,setClassId]=useState("");
 const [quizzes,setQuizzes]=useState<Quiz[]>([]);
 const [selected,setSelected]=useState("");
 const [detail,setDetail]=useState<QuizDetail|null>(null);
 const [results,setResults]=useState<Result[]>([]);
 const [answers,setAnswers]=useState<Record<string,number>>({});
 const [name,setName]=useState("");
 const [minutes,setMinutes]=useState(30);
 const [due,setDue]=useState("");
 const [prompt,setPrompt]=useState("");
 const [options,setOptions]=useState(["","","",""]);
 const [correct,setCorrect]=useState(0);
 const [message,setMessage]=useState("");
 const [error,setError]=useState("");
 const [busy,setBusy]=useState(false);
 const [now,setNow]=useState(Date.now());
 useEffect(()=>{const t=setInterval(()=>setNow(Date.now()),1000);return ()=>clearInterval(t)},[]);
 useEffect(()=>{if(!classes.some(c=>c.id===classId)){setClassId(classes[0]?.id||"");setSelected("");setDetail(null)}},[classes,classId]);
 const request=useCallback(async<T,>(path:string,method="GET",body?:object):Promise<T>=>{
  const response=await fetch("/api/quizzes/"+path,{
   method,cache:"no-store",
   headers:{"content-type":"application/json"},
   body:body?JSON.stringify(body):undefined
  });
  const data=await response.json() as T&{error?:string};
  if(!response.ok)throw new Error(data.error||"Không thể xử lý bài kiểm tra");
  return data;
 },[]);
 const load=useCallback(async(id:string)=>{
  if(!id){setQuizzes([]);return}
  try{const data=await request<{items:Quiz[]}>("classes/"+encodeURIComponent(id)+"/quizzes");
   setQuizzes(data.items);setError("");
  }catch(err){setError(err instanceof Error?err.message:"Không thể tải bài kiểm tra")}
 },[request]);
 useEffect(()=>{void load(classId)},[classId,load]);
 async function open(id:string){
  setBusy(true);setError("");setMessage("");
  try{
   const data=await request<QuizDetail>("quizzes/"+id);
   setSelected(id);setDetail(data);setAnswers({});
   if(staff){
    const result=await request<{items:Result[]}>("quizzes/"+id+"/results");
    setResults(result.items);
   }
  }catch(err){setError(err instanceof Error?err.message:"Không thể mở đề");}
  finally{setBusy(false)}
 }
 async function perform(path:string,body:object,messageText:string,reopen=false){
  setBusy(true);setError("");setMessage("");
  try{
   const result=await request<{id?:string;grade?:number;correct?:number;total?:number}>(path,"POST",body);
   setMessage(result.grade!==undefined?"Đã nộp. Điểm: "+result.grade+"/10 ("+result.correct+"/"+result.total+" câu đúng).":messageText);
   if(classId)await load(classId);
   if(result.id&&!reopen){setName("");await open(result.id)}
   else if(selected&&reopen)await open(selected);
   return result;
  }catch(err){setError(err instanceof Error?err.message:"Không thể thực hiện");return null}
  finally{setBusy(false)}
 }
 function create(event:FormEvent<HTMLFormElement>){
  event.preventDefault();
  if(!classId||!due)return;
  const date=new Date(due);
  if(!Number.isFinite(date.getTime())){setError("Hạn nộp không hợp lệ");return}
  void perform("classes/"+classId+"/quizzes",{title:name,duration_minutes:minutes,due_at:date.toISOString()},"Đã tạo bản nháp đề kiểm tra");
 }
 function addQuestion(event:FormEvent<HTMLFormElement>){
  event.preventDefault();
  if(!selected)return;
  void perform("quizzes/"+selected+"/questions",{prompt,options,correct_option:correct,position:(detail?.items.length||0)+1},"Đã thêm câu hỏi",true).then(r=>{
   if(r){setPrompt("");setOptions(["","","",""]);setCorrect(0);}
  });
 }
 function submit(event:FormEvent<HTMLFormElement>){
  event.preventDefault();if(!detail)return;
  if(detail.items.some(q=>answers[q.id]===undefined)){setError("Vui lòng trả lời tất cả câu hỏi.");return}
  void perform("quizzes/"+detail.id+"/submit",{answers},"Đã nộp bài",true);
 }
 const active=quizzes.find(q=>q.id===selected);
 const msLeft=detail?.expires_at?new Date(detail.expires_at).getTime()-now:0;
 const mins=Math.max(0,Math.ceil(msLeft/60000));
 return <div className="quiz-studio">
  <Card>
   <h2 className="icon-heading"><ClipboardList size={22} aria-hidden="true"/>Bài kiểm tra trắc nghiệm</h2>
   <p className="muted">Đề một lần làm, giới hạn thời gian trên máy chủ. Học sinh chỉ xem câu hỏi trong thời gian làm bài; đáp án được giữ kín.</p>
   {classes.length===0?<p className="muted">Chưa có lớp được phân công hoặc ghi danh.</p>:<>
    <label className="quiz-label">Lớp học
     <select value={classId} onChange={e=>{setClassId(e.target.value);setSelected("");setDetail(null)}}>{classes.map(cl=><option key={cl.id} value={cl.id}>{cl.name} · Khối {cl.grade}</option>)}</select>
    </label>
    <div className="quiz-list">
     {quizzes.map(q=><button type="button" key={q.id} className={"class-item "+(selected===q.id?"selected":"")} onClick={()=>void open(q.id)}>
      <strong>{q.title}</strong>
      <small><AlarmClock size={14} aria-hidden="true"/>{q.duration_minutes} phút · {q.questions} câu · {q.status==="draft"?"Bản nháp":"Đã phát hành"}{q.grade!==null?" · Điểm "+q.grade+"/10":""}</small>
     </button>)}
     {!quizzes.length&&<p className="muted">Chưa có bài kiểm tra cho lớp này.</p>}
    </div>
    {staff&&<form className="portal-form" onSubmit={create}>
     <h3><Plus size={17} aria-hidden="true"/>Tạo đề mới</h3>
     <label>Tên bài kiểm tra<input required maxLength={180} minLength={3} value={name} onChange={e=>setName(e.target.value)} placeholder="Kiểm tra chương 1"/></label>
     <label>Thời gian làm (phút)<input type="number" min="5" max="180" required value={minutes} onChange={e=>setMinutes(Number(e.target.value))}/></label>
     <label>Hạn kết thúc<input type="datetime-local" required value={due} onChange={e=>setDue(e.target.value)}/></label>
     <Button disabled={busy||!classId} type="submit"><Plus size={17} aria-hidden="true"/>Tạo bản nháp</Button>
    </form>}
   </>}
  </Card>
  {message&&<p className="portal-message" role="status">{message}</p>}
  {error&&<p className="portal-message" role="alert">{error}</p>}
  {detail&&<Card>
   <div className="quiz-title">
    <h2 className="icon-heading"><FileQuestion size={22} aria-hidden="true"/>{active?.title||"Đề kiểm tra"}</h2>
    <span className="pill">{detail.submitted?"Đã nộp":detail.status==="draft"?"Đang soạn":detail.started?"Đang làm bài":"Chưa bắt đầu"}</span>
   </div>
   <p className="muted">Hạn kết thúc: {new Date(detail.due_at).toLocaleString("vi-VN")}</p>
   {detail.submitted&&<div className="quiz-score"><CheckCircle2 size={22} aria-hidden="true"/>Đã hoàn thành · Điểm {detail.grade??"chưa có"}/10</div>}
   {staff&&detail.status==="draft"&&<div className="portal-form">
    <h3>Soạn câu hỏi</h3>
    {detail.items.map(q=><div className="quiz-question" key={q.id}><strong>{q.position}. {q.prompt}</strong>
     {q.options.map((v,i)=><p key={i}>{String.fromCharCode(65+i)}. {v}{i===q.correct_option?" (Đáp án)":""}</p>)}</div>)}
    <form className="portal-form" onSubmit={addQuestion}>
     <label>Nội dung câu hỏi<textarea rows={3} required minLength={3} maxLength={2000} value={prompt} onChange={e=>setPrompt(e.target.value)}/></label>
     {options.map((value,i)=><label key={i}>Phương án {String.fromCharCode(65+i)}
      <input required maxLength={500} value={value} onChange={e=>setOptions(v=>v.map((x,j)=>j===i?e.target.value:x))}/></label>)}
     <label>Đáp án chính xác<select value={correct} onChange={e=>setCorrect(Number(e.target.value))}>
      {options.map((_,i)=><option value={i} key={i}>Đáp án {String.fromCharCode(65+i)}</option>)}
     </select></label>
     <Button disabled={busy||detail.items.length>=100} type="submit"><Plus size={16} aria-hidden="true"/>Thêm câu hỏi</Button>
    </form>
    <Button disabled={busy||detail.items.length===0} onClick={()=>void perform("quizzes/"+selected+"/publish",{},"Đã phát hành đề; không thể sửa đáp án.",true)}>
     <ShieldCheck size={17} aria-hidden="true"/>Duyệt và phát hành
    </Button>
   </div>}
   {staff&&detail.status==="published"&&<>
    <p className="muted">Đề đã phát hành. Nội dung và đáp án được khóa để bảo đảm kết quả chấm điểm.</p>
    {detail.items.map(q=><div className="quiz-question" key={q.id}><strong>{q.position}. {q.prompt}</strong>{q.options.map((v,i)=><p key={i}>{String.fromCharCode(65+i)}. {v}{i===q.correct_option?" (Đáp án)":""}</p>)}</div>)}
    <h3 className="icon-heading"><UsersRound size={18} aria-hidden="true"/>Kết quả lớp</h3>
    {results.map(r=><div className="portal-row" key={r.student_id}><strong>{r.name}</strong><span>{r.submitted?String(r.grade)+"/10":"Đang làm"}</span></div>)}
    {!results.length&&<p className="muted">Chưa có học sinh bắt đầu.</p>}
   </>}
   {!staff&&!detail.submitted&&!detail.started&&<Button disabled={busy||new Date(detail.due_at).getTime()<=now} onClick={()=>void perform("quizzes/"+selected+"/start",{},"Đã bắt đầu tính giờ",true)}>
    <AlarmClock size={17} aria-hidden="true"/>Bắt đầu làm bài
   </Button>}
   {!staff&&detail.started&&!detail.submitted&&<form className="portal-form" onSubmit={submit}>
    {msLeft>0&&detail.items.length>0?<><p className="quiz-clock"><AlarmClock size={17} aria-hidden="true"/>Còn khoảng {mins} phút · Thời gian được kiểm soát ở máy chủ</p>
     {detail.items.map(q=><fieldset className="quiz-question" key={q.id}><legend><strong>{q.position}. {q.prompt}</strong></legend>
      {q.options.map((option,index)=><label key={index} className="quiz-option">
       <input type="radio" name={"question-"+q.id} checked={answers[q.id]===index} onChange={()=>setAnswers(a=>({...a,[q.id]:index}))}/>
       <span>{String.fromCharCode(65+index)}. {option}</span>
      </label>)}
     </fieldset>)}
     <Button type="submit" disabled={busy||detail.items.some(q=>answers[q.id]===undefined)}><Send size={17} aria-hidden="true"/>Nộp bài và nhận điểm</Button>
    </>:<p className="muted">Hết giờ. Hệ thống không nhận bài nộp sau hạn; liên hệ giáo viên nếu cần hỗ trợ.</p>}
   </form>}
  </Card>}
 </div>;
}