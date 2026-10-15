"use client";

import { FormEvent, useCallback, useEffect, useState } from "react";
import { School, UsersRound, GraduationCap, Bell, LayoutDashboard, ClipboardCheck, CalendarDays, NotebookPen, CalendarClock, CircleCheck, CircleX, MessageCircle, BarChart3 } from "lucide-react";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import MessagesPanel from "./messages-panel";
import ReportsPanel from "./reports-panel";
type RecordItem=Record<string,unknown>;
type Overview={classes:number;students:number;teachers:number;announcements:number};
type Year={id:string;name:string;starts_on:string;ends_on:string};
type ClassRoom={id:string;name:string;grade:number;academic_year_id:string;academic_year:string};
type Notice={id:string;title:string;content:string;audience:string;created_at:string};
type Student={id:string;name:string};
type Attendance={student_id:string;status:string;note:string};
type Assignment={id:string;title:string;instructions:string;due_at?:string};
type Leave={id:string;student_id:string;reason:string;start_on:string;end_on:string;status:string};
type Period={id:string;weekday:number;period:number;subject_code:string;room:string};

const day=new Date().toLocaleDateString("en-CA",{timeZone:"Asia/Ho_Chi_Minh"});
const weekdays=["","Thứ 2","Thứ 3","Thứ 4","Thứ 5","Thứ 6","Thứ 7","Chủ nhật"];

const uid=(v:string)=>/^[a-f0-9-]{36}$/i.test(v);

export default function SchoolPanel({schoolId,userRoles}:{schoolId:string;userRoles:string[]}){
 const isAdmin=userRoles.some(r=>r==="school_admin"||r==="principal");
 const staff=isAdmin||userRoles.includes("teacher");
 const [tab,setTab]=useState("overview");
 const [loading,setLoading]=useState(true);
 const [message,setMessage]=useState("");
 const [busy,setBusy]=useState(false);
 const [years,setYears]=useState<Year[]>([]);
 const [classes,setClasses]=useState<ClassRoom[]>([]);
 const [notices,setNotices]=useState<Notice[]>([]);
 const [overview,setOverview]=useState<Overview>({classes:0,students:0,teachers:0,announcements:0});
 const [classId,setClassId]=useState("");
 const [roster,setRoster]=useState<Student[]>([]);
 const [attendance,setAttendance]=useState<Attendance[]>([]);
 const [assignments,setAssignments]=useState<Assignment[]>([]);
 const [periods,setPeriods]=useState<Period[]>([]);
 const [leaves,setLeaves]=useState<Leave[]>([]);
 const [selectedAssignment,setSelectedAssignment]=useState("");
 const [submissions,setSubmissions]=useState<RecordItem[]>([]);
 const [date,setDate]=useState(day);
 const [yearName,setYearName]=useState("");
 const [yearStart,setYearStart]=useState("");
 const [yearEnd,setYearEnd]=useState("");
 const [className,setClassName]=useState("");
 const [yearId,setYearId]=useState("");
 const [grade,setGrade]=useState(6);
 const [newStudent,setNewStudent]=useState("");
 const [teacherId,setTeacherId]=useState("");
 const [guardianId,setGuardianId]=useState("");
 const [guardianStudent,setGuardianStudent]=useState("");
 const [subject,setSubject]=useState("toan");
 const [noticeTitle,setNoticeTitle]=useState("");
 const [noticeContent,setNoticeContent]=useState("");
 const [audience,setAudience]=useState("all");
 const [assignmentTitle,setAssignmentTitle]=useState("");
 const [instructions,setInstructions]=useState("");
 const [work,setWork]=useState("");
 const [score,setScore]=useState("");
 const [feedback,setFeedback]=useState("");
 const [studentToGrade,setStudentToGrade]=useState("");
 const [leaveReason,setLeaveReason]=useState("");
 const [leaveEnd,setLeaveEnd]=useState(day);
 const [slotDay,setSlotDay]=useState(1);
 const [slotPeriod,setSlotPeriod]=useState(1);
 const [slotRoom,setSlotRoom]=useState("");
 const url=(resource:"schools"|"classes"|"assignments",id:string,path:string)=>"/api/school/"+resource+"/"+encodeURIComponent(id)+"/"+path;
 const request=useCallback(async <T,>(path:string,method="GET",data?:RecordItem):Promise<T>=>{
   const res=await fetch(path,{method,cache:"no-store",headers:{"content-type":"application/json"},body:data?JSON.stringify(data):undefined});
   const payload=await res.json() as T & {error?:string};
   if(!res.ok)throw new Error(payload.error||"Request failed");
   return payload;
 },[]);
 const load=useCallback(async()=>{
  try{
   const [y,c,n,o]=await Promise.all([
    request<{items:Year[]}>(url("schools",schoolId,"years")),
    request<{items:ClassRoom[]}>(url("schools",schoolId,"classes")),
    request<{items:Notice[]}>(url("schools",schoolId,"announcements")),
    request<Overview>(url("schools",schoolId,"overview"))]);
   setYears(y.items);setClasses(c.items);setNotices(n.items);setOverview(o);
  }catch(err){setMessage(err instanceof Error?err.message:"Lỗi tải dữ liệu");}
  finally{setLoading(false);}
 },[request,schoolId]);
 const loadClass=useCallback(async(id:string,currentDate:string)=>{
  if(!id)return;
  try{
   const [a,b,c,d,e]=await Promise.all([
    staff?request<{items:Student[]}>(url("classes",id,"students")):Promise.resolve({items:[]}),
    staff?request<{items:Attendance[]}>(url("classes",id,"attendance")+"?date="+currentDate):Promise.resolve({items:[]}),
    request<{items:Assignment[]}>(url("classes",id,"assignments")),
    request<{items:Period[]}>(url("classes",id,"timetable")),
    request<{items:Leave[]}>(url("classes",id,"leave-requests")).catch(()=>({items:[]}))
   ]);
   setRoster(a.items);setAttendance(b.items);setAssignments(c.items);setPeriods(d.items);setLeaves(e.items);
  }catch(err){setMessage(err instanceof Error?err.message:"Lỗi tải dữ liệu lớp học");}
 },[request,staff]);
 useEffect(()=>{void load();},[load]);
 useEffect(()=>{if(classId){void loadClass(classId,date);}},[classId,date,loadClass]);
 async function submit(path:string,method:string,payload:RecordItem,ok:string,after?:()=>void){
  setBusy(true);setMessage("");
  try{await request(path,method,payload);setMessage(ok);after?.();await load();if(classId)await loadClass(classId,date);}
  catch(err){setMessage(err instanceof Error?err.message:"Không thể lưu dữ liệu");}
  finally{setBusy(false);}
 }
 async function attendanceSet(student:string,status:string){
  await submit(url("classes",classId,"attendance"),"PUT",{student_id:student,date,status,note:""},"Đã lưu điểm danh");
 }
 async function loadSubmissions(aid:string){
  setSelectedAssignment(aid);
  try{const data=await request<{items:RecordItem[]}>(url("assignments",aid,"submissions"));setSubmissions(data.items);}
  catch(err){setMessage(err instanceof Error?err.message:"Không thể đọc bài nộp");}
 }
 const handle=(event:FormEvent<HTMLFormElement>,path:string,method:string,data:RecordItem,success:string,reset?:()=>void)=>{
  event.preventDefault();void submit(path,method,data,success,reset);
 };

 const panel = (<>
  {message&&<div role="status" className="portal-message">{message}</div>}
  {loading?<div className="glass-card" role="status">Đang tải dữ liệu trường học…</div>:null}
  {tab==="overview"&&<div className="glass-grid">
   {[
    { name:"Lớp học", value:overview.classes, Icon:School },
    { name:"Học sinh", value:overview.students, Icon:UsersRound },
    { name:"Giáo viên", value:overview.teachers, Icon:GraduationCap },
    { name:"Thông báo", value:overview.announcements, Icon:Bell },
   ].map(({name,value,Icon})=><div className="glass-card stat-card" key={name}><span><Icon size={24} strokeWidth={1.9} aria-hidden="true"/></span><strong>{value}</strong><small>{name}</small></div>)}
   <section className="glass-card glass-wide"><h2>Năm học</h2><p className="muted">Chọn năm học khi tạo lớp. Mỗi trường có dữ liệu độc lập.</p>
    {years.length?years.map(y=><div key={y.id} className="portal-row"><strong>{y.name}</strong><span>{y.starts_on} → {y.ends_on}</span></div>):<p>Chưa có năm học.</p>}
    {isAdmin&&<form className="portal-form" onSubmit={e=>handle(e,url("schools",schoolId,"years"),"POST",{name:yearName,starts_on:yearStart,ends_on:yearEnd},"Đã tạo năm học",()=>{setYearName("");setYearStart("");setYearEnd("");})}>
     <h3>Thêm năm học</h3><label>Tên năm học<input required placeholder="2026–2027" value={yearName} onChange={e=>setYearName(e.target.value)}/></label>
     <label>Ngày bắt đầu<input type="date" required value={yearStart} onChange={e=>setYearStart(e.target.value)}/></label>
     <label>Ngày kết thúc<input type="date" required value={yearEnd} onChange={e=>setYearEnd(e.target.value)}/></label>
     <button disabled={busy}>Tạo năm học</button>
    </form>}
   </section>
  </div>}
  {["classes","learning","attendance","schedule","leaves"].includes(tab)&&<>
   <div className="glass-card class-picker">
    <label>Chọn lớp <select value={classId} onChange={e=>{setClassId(e.target.value);setSelectedAssignment("");}}>
     <option value="">— Chọn một lớp —</option>
     {classes.map(c=><option key={c.id} value={c.id}>{c.name} · Lớp {c.grade} · {c.academic_year}</option>)}
    </select></label>
    <span>{classes.length} lớp được thiết lập</span>
   </div>
   {tab==="classes"&&<div className="portal-two">
    <section className="glass-card">
     <h2>Quản lý lớp</h2>
     {classes.map(c=><button key={c.id} className={"class-item "+(classId===c.id?"selected":"")} onClick={()=>setClassId(c.id)}>{c.name}<small>Lớp {c.grade} · {c.academic_year}</small></button>)}
     {!classes.length&&<p className="muted">Chưa có lớp học.</p>}
     {isAdmin&&<form className="portal-form" onSubmit={e=>handle(e,url("schools",schoolId,"classes"),"POST",{name:className,academic_year_id:yearId,grade},"Đã tạo lớp",()=>setClassName(""))}>
      <h3>Thêm lớp học</h3><label>Niên khóa<select required value={yearId} onChange={e=>setYearId(e.target.value)}><option value="">Chọn năm học</option>{years.map(y=><option key={y.id} value={y.id}>{y.name}</option>)}</select></label>
      <label>Tên lớp<input required value={className} onChange={e=>setClassName(e.target.value)} placeholder="6A1"/></label>
      <label>Khối<select value={grade} onChange={e=>setGrade(Number(e.target.value))}>{Array.from({length:12},(_,i)=><option key={i} value={i+1}>Lớp {i+1}</option>)}</select></label>
      <button disabled={busy||!yearId}>Tạo lớp</button>
     </form>}
    </section>
    <section className="glass-card"><h2>Danh sách học sinh</h2>
     {classId? <>{staff?roster.map(s=><div className="portal-row" key={s.id}><strong>{s.name}</strong><small>{s.id.slice(0,8)}</small></div>):<p className="muted">Danh sách riêng tư chỉ hiện với nhân sự có quyền.</p>}
      {!roster.length&&staff&&<p className="muted">Lớp chưa có học sinh.</p>}
      {isAdmin&&<><form className="portal-form" onSubmit={e=>handle(e,url("classes",classId,"students"),"POST",{student_id:newStudent},"Đã thêm học sinh vào lớp",()=>setNewStudent(""))}>
       <h3>Ghi danh</h3><p className="muted">Học sinh phải có tài khoản trong trường trước.</p><label>ID học sinh<input value={newStudent} onChange={e=>setNewStudent(e.target.value)} required placeholder="UUID tài khoản học sinh"/></label><button disabled={busy||!uid(newStudent)}>Thêm vào lớp</button>
      </form>
      <form className="portal-form" onSubmit={e=>handle(e,url("classes",classId,"teachers"),"POST",{teacher_id:teacherId,subject_code:subject},"Đã phân công giáo viên",()=>setTeacherId(""))}>
       <h3>Phân công giáo viên</h3><label>ID giáo viên<input required value={teacherId} onChange={e=>setTeacherId(e.target.value)}/></label><label>Mã môn học<input required value={subject} onChange={e=>setSubject(e.target.value)}/></label><button disabled={busy||!uid(teacherId)}>Phân công</button>
      </form>
      <form className="portal-form" onSubmit={e=>handle(e,url("schools",schoolId,"guardians"),"POST",{guardian_id:guardianId,student_id:guardianStudent},"Đã xác minh liên kết phụ huynh — học sinh",()=>{setGuardianId("");setGuardianStudent("");})}>
       <h3>Xác minh phụ huynh</h3>
       <p className="muted">Chỉ xác minh sau khi nhà trường đối chiếu giấy tờ hoặc quan hệ người giám hộ.</p>
       <label>ID phụ huynh<input value={guardianId} onChange={e=>setGuardianId(e.target.value)} required placeholder="UUID tài khoản phụ huynh"/></label>
       <label>ID học sinh<input value={guardianStudent} onChange={e=>setGuardianStudent(e.target.value)} required placeholder="UUID tài khoản học sinh"/></label>
       <button disabled={busy||!uid(guardianId)||!uid(guardianStudent)}>Xác minh quan hệ</button>
      </form></>}
     </>:<p>Chọn lớp để xem danh sách.</p>}
    </section>
   </div>}
   {tab==="attendance"&&<section className="glass-card"><h2 className="icon-heading"><ClipboardCheck size={23} strokeWidth={1.9} aria-hidden="true"/>Điểm danh</h2>
    <label className="day-control">Ngày <input type="date" value={date} onChange={e=>setDate(e.target.value)}/></label>
    {!classId?<p className="muted">Chọn lớp để ghi nhận điểm danh.</p>:roster.map(st=>{
     const current=attendance.find(a=>a.student_id===st.id)?.status||"";
     return <div className="attendance-row" key={st.id}><strong>{st.name}</strong><select aria-label={"Điểm danh "+st.name} value={current} disabled={busy||!staff} onChange={e=>void attendanceSet(st.id,e.target.value)}><option value="">Chưa điểm danh</option><option value="present">Có mặt</option><option value="absent">Vắng</option><option value="late">Đi muộn</option><option value="excused">Có phép</option></select></div>
    })}
    {classId&&!roster.length&&<p className="muted">Chưa có học sinh hoặc bạn không có quyền xem danh sách.</p>}
   </section>}
   {tab==="schedule"&&<section className="glass-card"><h2 className="icon-heading"><CalendarDays size={23} strokeWidth={1.9} aria-hidden="true"/>Thời khóa biểu</h2>
    {periods.length?<div className="schedule-grid">{periods.map(p=><div className="schedule-cell" key={p.id}><small>{weekdays[p.weekday]} · Tiết {p.period}</small><strong>{p.subject_code}</strong><span>{p.room||"—"}</span></div>)}</div>:<p className="muted">Chưa có lịch học.</p>}
    {classId&&isAdmin&&<form className="portal-form" onSubmit={e=>handle(e,url("classes",classId,"timetable"),"POST",{weekday:slotDay,period:slotPeriod,subject_code:subject,room:slotRoom},"Đã lưu tiết học")}>
     <h3>Thêm hoặc chỉnh tiết học</h3><label>Thứ<select value={slotDay} onChange={e=>setSlotDay(Number(e.target.value))}>{weekdays.slice(1).map((name,i)=><option key={i} value={i+1}>{name}</option>)}</select></label>
     <label>Tiết<input type="number" value={slotPeriod} min={1} max={15} onChange={e=>setSlotPeriod(Number(e.target.value))}/></label>
     <label>Môn học<input required value={subject} onChange={e=>setSubject(e.target.value)}/></label><label>Phòng<input value={slotRoom} onChange={e=>setSlotRoom(e.target.value)}/></label>
     <button disabled={busy}>Lưu thời khóa biểu</button>
    </form>}
   </section>}
   {tab==="learning"&&<div className="portal-two">
    <section className="glass-card"><h2 className="icon-heading"><NotebookPen size={23} strokeWidth={1.9} aria-hidden="true"/>Bài tập của lớp</h2>
     {assignments.map(a=><button key={a.id} className={"class-item "+(selectedAssignment===a.id?"selected":"")} onClick={()=>void loadSubmissions(a.id)}>{a.title}<small>{a.due_at?"Hạn "+new Date(a.due_at).toLocaleDateString("vi-VN"):"Không giới hạn hạn nộp"}</small></button>)}
     {!assignments.length&&<p className="muted">Chưa có bài tập.</p>}
     {classId&&staff&&<form className="portal-form" onSubmit={e=>handle(e,url("classes",classId,"assignments"),"POST",{title:assignmentTitle,instructions},"Đã tạo bài tập",()=>{setAssignmentTitle("");setInstructions("");})}>
      <h3>Giao bài tập</h3><label>Tiêu đề<input required value={assignmentTitle} onChange={e=>setAssignmentTitle(e.target.value)}/></label><label>Hướng dẫn<textarea value={instructions} onChange={e=>setInstructions(e.target.value)} rows={4}/></label><button disabled={busy}>Giao bài</button>
     </form>}
    </section>
    <section className="glass-card"><h2>Bài nộp và đánh giá</h2>
     {!selectedAssignment?<p className="muted">Chọn bài tập để xem hoặc nộp bài.</p>:<>
      {submissions.map((x,i)=><div key={i} className="submission-card"><strong>{String(x.name||"Bài nộp")}</strong><p>{String(x.content||"")}</p><small>{x.grade==null?"Chưa chấm":"Điểm: "+String(x.grade)+"/10"}</small></div>)}
      {userRoles.includes("student")&&<form className="portal-form" onSubmit={e=>handle(e,url("assignments",selectedAssignment,"submissions"),"POST",{content:work},"Đã nộp bài",()=>{setWork("");void loadSubmissions(selectedAssignment);})}><label>Nội dung bài nộp<textarea required rows={5} value={work} onChange={e=>setWork(e.target.value)}/></label><button disabled={busy}>Nộp bài</button></form>}
      {staff&&<form className="portal-form" onSubmit={e=>handle(e,url("assignments",selectedAssignment,"submissions"),"PATCH",{student_id:studentToGrade,grade:Number(score),feedback},"Đã lưu điểm",()=>void loadSubmissions(selectedAssignment))}>
       <h3>Chấm điểm</h3><label>ID học sinh<input required value={studentToGrade} onChange={e=>setStudentToGrade(e.target.value)}/></label>
       <label>Điểm (0–10)<input type="number" step=".25" min="0" max="10" value={score} required onChange={e=>setScore(e.target.value)}/></label>
       <label>Nhận xét<textarea value={feedback} onChange={e=>setFeedback(e.target.value)}/></label><button disabled={busy||!uid(studentToGrade)}>Lưu điểm</button>
      </form>}
     </>}
    </section>
   </div>}
   {tab==="leaves"&&<section className="glass-card"><h2 className="icon-heading"><CalendarClock size={23} strokeWidth={1.9} aria-hidden="true"/>Đơn xin nghỉ</h2>
    {leaves.map(l=><div className="portal-row" key={l.id}><div><strong>{l.reason}</strong><small>{l.start_on} → {l.end_on} · {l.status}</small></div>
     {staff&&l.status==="pending"&&<div className="portal-inline"><button onClick={()=>void submit(url("classes",classId,"leave-requests"),"PATCH",{request_id:l.id,status:"approved"},"Đã duyệt đơn")}><CircleCheck size={15} aria-hidden="true"/>Duyệt</button><button onClick={()=>void submit(url("classes",classId,"leave-requests"),"PATCH",{request_id:l.id,status:"rejected"},"Đã từ chối đơn")}><CircleX size={15} aria-hidden="true"/>Từ chối</button></div>}
    </div>)}
    {classId&&userRoles.includes("student")&&<form className="portal-form" onSubmit={e=>handle(e,url("classes",classId,"leave-requests"),"POST",{start_on:date,end_on:leaveEnd,reason:leaveReason},"Đã gửi đơn xin nghỉ",()=>setLeaveReason(""))}>
     <h3>Gửi đơn</h3><label>Từ ngày<input type="date" value={date} onChange={e=>setDate(e.target.value)}/></label>
     <label>Đến ngày<input type="date" value={leaveEnd} onChange={e=>setLeaveEnd(e.target.value)}/></label>
     <label>Lý do<textarea required rows={3} value={leaveReason} onChange={e=>setLeaveReason(e.target.value)}/></label><button disabled={busy}>Gửi đơn</button>
    </form>}
    {!leaves.length&&<p className="muted">Chưa có đơn xin nghỉ.</p>}
   </section>}
  </>}
  {tab==="messages"&&<MessagesPanel/>}
  {tab==="reports"&&isAdmin&&<ReportsPanel schoolId={schoolId}/>}
  {tab==="announcements"&&<section className="glass-card"><h2 className="icon-heading"><Bell size={23} strokeWidth={1.9} aria-hidden="true"/>Thông báo nhà trường</h2>
   {notices.map(n=><article className="notice" key={n.id}><span className="pill">{n.audience}</span><h3>{n.title}</h3><p>{n.content}</p><small>{new Date(n.created_at).toLocaleDateString("vi-VN")}</small></article>)}
   {!notices.length&&<p className="muted">Chưa có thông báo.</p>}
   {isAdmin&&<form className="portal-form" onSubmit={e=>handle(e,url("schools",schoolId,"announcements"),"POST",{title:noticeTitle,content:noticeContent,audience},"Đã đăng thông báo",()=>{setNoticeTitle("");setNoticeContent("");})}>
    <h3>Đăng thông báo</h3><label>Tiêu đề<input required value={noticeTitle} onChange={e=>setNoticeTitle(e.target.value)}/></label>
    <label>Nội dung<textarea required rows={4} value={noticeContent} onChange={e=>setNoticeContent(e.target.value)}/></label>
    <label>Đối tượng<select value={audience} onChange={e=>setAudience(e.target.value)}><option value="all">Toàn trường</option><option value="teacher">Giáo viên</option><option value="student">Học sinh</option><option value="guardian">Phụ huynh</option></select></label>
    <button disabled={busy}>Đăng thông báo</button>
   </form>}
  </section>}
  <p className="portal-footnote">Không gian dữ liệu của trường được bảo vệ bằng phân quyền. Các tài khoản và học liệu chỉ được cấp phát bởi đơn vị có thẩm quyền.</p>

 </>);
 const nav = [
  { id:"overview", label:"Tổng quan", Icon:LayoutDashboard },
  { id:"classes", label:"Lớp học", Icon:School },
  { id:"learning", label:"Bài tập", Icon:NotebookPen },
  { id:"attendance", label:"Điểm danh", Icon:ClipboardCheck },
  { id:"schedule", label:"Thời khóa biểu", Icon:CalendarDays },
  { id:"announcements", label:"Thông báo", Icon:Bell },
  { id:"leaves", label:"Xin nghỉ", Icon:CalendarClock },
  ...(userRoles.includes("teacher")?[{id:"messages",label:"Tin nhắn",Icon:MessageCircle}]:[]),
  ...(isAdmin?[{id:"reports",label:"Báo cáo",Icon:BarChart3}]:[]),
 ];
 return <Tabs className="school-portal" value={tab} onValueChange={setTab}>
  <TabsList aria-label="Phân hệ trường học">
   {nav.map(({id,label,Icon})=><TabsTrigger value={id} key={id} icon={Icon}>{label}</TabsTrigger>)}
  </TabsList>
  {nav.map(({id})=><TabsContent key={id} value={id} forceMount>{tab===id?panel:null}</TabsContent>)}
 </Tabs>;
}