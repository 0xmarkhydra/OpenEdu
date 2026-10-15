"use client";
import { FormEvent,useEffect,useState } from "react";
type Child={student_id:string;name:string;class_id:string;class_name:string};
type Grade={title:string;class_name:string;grade:number|null;feedback:string;submitted_at:string};
type Attendance={date:string;status:string;class_name:string;note:string};
const today=new Date().toLocaleDateString("en-CA",{timeZone:"Asia/Ho_Chi_Minh"});
export default function ParentPanel(){
 const [children,setChildren]=useState<Child[]>([]);
 const [selected,setSelected]=useState("");
 const [grades,setGrades]=useState<Grade[]>([]);
 const [attendance,setAttendance]=useState<Attendance[]>([]);
 const [status,setStatus]=useState("");
 const [busy,setBusy]=useState(false);
 const [from,setFrom]=useState(today);
 const [to,setTo]=useState(today);
 const [reason,setReason]=useState("");
 useEffect(()=>{
  let alive=true;
  fetch("/api/guardian/children",{cache:"no-store"}).then(async res=>{
   const data=await res.json();
   if(!res.ok)throw new Error(data.error||"Không thể tải thông tin");
   if(alive){setChildren(data.items||[]);setSelected(data.items?.[0]?.student_id||"");}
  }).catch(err=>{if(alive)setStatus(err.message);});
  return()=>{alive=false};
 },[]);
 useEffect(()=>{
  if(!selected)return;
  let alive=true;
  Promise.all(["grades","attendance"].map(async section=>{
   const res=await fetch("/api/guardian/children/"+encodeURIComponent(selected)+"/"+section,{cache:"no-store"});
   const data=await res.json();
   if(!res.ok)throw new Error(data.error||"Không thể tải dữ liệu");
   return data.items||[];
  })).then(([g,a])=>{if(alive){setGrades(g);setAttendance(a)}}).catch(err=>{if(alive)setStatus(err.message)});
  return()=>{alive=false};
 },[selected]);
 const child=children.find(c=>c.student_id===selected);
 async function submit(event:FormEvent<HTMLFormElement>){
  event.preventDefault();if(!child?.class_id)return;
  setBusy(true);setStatus("");
  try{
   const response=await fetch("/api/guardian/children/"+encodeURIComponent(selected)+"/leave",{method:"POST",headers:{"content-type":"application/json"},body:JSON.stringify({class_id:child.class_id,start_on:from,end_on:to,reason})});
   const data=await response.json();
   if(!response.ok)throw new Error(data.error||"Gửi đơn thất bại");
   setStatus("Đã gửi đơn xin nghỉ đến nhà trường.");setReason("");
  }catch(err){setStatus(err instanceof Error?err.message:"Gửi đơn thất bại")}
  finally{setBusy(false)}
 }
 return <section className="parent-portal school-portal">
  <section className="glass-card parent-heading">
   <span className="eyebrow section-eyebrow">KẾT NỐI GIA ĐÌNH</span>
   <h2>Sổ liên lạc điện tử</h2>
   <p className="muted">Theo dõi học tập và gửi yêu cầu đến nhà trường. Chỉ học sinh đã xác minh quan hệ được hiển thị.</p>
   <label className="portal-select">Chọn học sinh<select value={selected} onChange={e=>setSelected(e.target.value)}>{children.length===0?<option value="">Chưa có liên kết đã xác minh</option>:children.map(c=><option key={c.student_id} value={c.student_id}>{c.name} · {c.class_name||"Chưa xếp lớp"}</option>)}</select></label>
  </section>
  {status&&<div role="status" className="portal-message">{status}</div>}
  {selected&&<div className="portal-two">
   <section className="glass-card"><h2>Kết quả học tập</h2>
    {grades.length?grades.map((g,i)=><div key={i} className="portal-row"><div><strong>{g.title}</strong><small>{g.class_name} · {g.feedback}</small></div><strong>{g.grade===null?"Chưa chấm":g.grade+"/10"}</strong></div>):<p className="muted">Chưa có kết quả được ghi nhận.</p>}
   </section>
   <section className="glass-card"><h2>Điểm danh gần đây</h2>
    {attendance.length?attendance.map((a,i)=><div key={i} className="portal-row"><div><strong>{a.date}</strong><small>{a.class_name} · {a.note}</small></div><span className="pill">{a.status}</span></div>):<p className="muted">Chưa có dữ liệu điểm danh.</p>}
   </section>
   {child?.class_id&&<section className="glass-card"><h2>Xin nghỉ học</h2><p className="muted">Đơn sẽ được giáo viên hoặc nhà trường xét duyệt.</p>
    <form className="portal-form" onSubmit={submit}>
     <label>Từ ngày<input type="date" value={from} onChange={e=>setFrom(e.target.value)} required/></label>
     <label>Đến ngày<input type="date" min={from} value={to} onChange={e=>setTo(e.target.value)} required/></label>
     <label>Lý do<textarea rows={4} required minLength={3} maxLength={1200} value={reason} onChange={e=>setReason(e.target.value)}/></label>
     <button disabled={busy}>{busy?"Đang gửi...":"Gửi đơn xin nghỉ"}</button>
    </form>
   </section>}
  </div>}
 </section>;
}
