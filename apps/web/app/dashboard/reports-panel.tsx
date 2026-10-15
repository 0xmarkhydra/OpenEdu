"use client";
import {useEffect,useState} from "react";
import {BarChart3,CalendarCheck,GraduationCap,NotebookPen,School,AlertCircle,RefreshCw,FileQuestion,ClipboardList} from "lucide-react";
import {Button} from "@/components/ui/button";
import {Card} from "@/components/ui/card";
type SchoolReport={
 summary:{total_students:number;total_classes:number;attendance_present_30d:number;attendance_total_30d:number;assignments:number;submissions:number;graded:number;pending_leave:number;quizzes:number;quiz_attempts:number;quiz_average_grade:number|null};
 classes:{id:string;name:string;grade:number;students:number}[];
 period:string;
};
export default function ReportsPanel({schoolId}:{schoolId:string}){
 const [report,setReport]=useState<SchoolReport|null>(null);
 const [error,setError]=useState("");
 const [loading,setLoading]=useState(true);
 async function load(){
  setLoading(true);setError("");
  try{
   const res=await fetch("/api/school/schools/"+encodeURIComponent(schoolId)+"/reports",{cache:"no-store"});
   const data=await res.json();
   if(!res.ok)throw new Error(data.error||"Không thể tải báo cáo");
   setReport(data as SchoolReport);
  }catch(e){setError(e instanceof Error?e.message:"Lỗi tải báo cáo");}
  finally{setLoading(false);}
 }
 useEffect(()=>{void load()},[schoolId]);
 const s=report?.summary;
 const rate=s&&s.attendance_total_30d?Math.round(s.attendance_present_30d*100/s.attendance_total_30d):null;
 const cards=s?[
  {label:"Học sinh",value:s.total_students,Icon:GraduationCap},
  {label:"Lớp học",value:s.total_classes,Icon:School},
  {label:"Bài tập",value:s.assignments,Icon:NotebookPen},
  {label:"Tỷ lệ có mặt (30 ngày)",value:rate===null?"Chưa có dữ liệu":rate+"%",Icon:CalendarCheck},
  {label:"Bài đã nộp",value:s.submissions,Icon:NotebookPen},
  {label:"Bài đã chấm",value:s.graded,Icon:BarChart3},
  {label:"Đề kiểm tra",value:s.quizzes,Icon:FileQuestion},
  {label:"Bài thi đã nộp",value:s.quiz_attempts,Icon:ClipboardList},
  {label:"Điểm thi trung bình",value:s.quiz_average_grade===null?"Chưa có":s.quiz_average_grade.toFixed(2)+"/10",Icon:BarChart3},
  {label:"Đơn chờ duyệt",value:s.pending_leave,Icon:AlertCircle}
 ]:[];
 return <section className="school-report">
  <div className="messages-heading">
   <div><h2 className="icon-heading"><BarChart3 size={24} aria-hidden="true"/>Báo cáo tổng hợp</h2>
    <p className="muted">Số liệu tổng hợp trong phạm vi trường. Không công khai thông tin cá nhân học sinh.</p>
   </div>
   <Button variant="outline" size="icon" aria-label="Làm mới báo cáo" onClick={()=>void load()}><RefreshCw size={18} aria-hidden="true"/></Button>
  </div>
  {error&&<div role="alert" className="portal-message">{error}</div>}
  {loading&&<p role="status">Đang cập nhật báo cáo…</p>}
  {s&&<><div className="glass-grid">{cards.map(({label,value,Icon})=><Card className="stat-card" key={label}>
   <span><Icon size={24} strokeWidth={1.9} aria-hidden="true"/></span><strong>{value}</strong><small>{label}</small>
  </Card>)}</div>
   <Card><h3>Quy mô lớp</h3>
    {report?.classes.length?report.classes.map(cls=><div key={cls.id} className="portal-row">
      <strong>{cls.name} · Khối {cls.grade}</strong><span>{cls.students} học sinh</span>
    </div>):<p className="muted">Chưa có lớp được tạo.</p>}
    <p className="portal-footnote">Tỷ lệ có mặt tính trên các bản ghi điểm danh trong 30 ngày gần nhất; không phải tỷ lệ trên toàn bộ lịch học.</p>
   </Card>
  </>}
 </section>;
}