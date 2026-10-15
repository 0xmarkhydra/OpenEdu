"use client";
import { FormEvent, useCallback, useEffect, useState } from "react";
import { CheckCheck, MessageCircle, RefreshCw, Send } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";

type Contact={class_id:string;recipient_id:string;name:string;class_name:string};
type Message={id:string;class_id:string;sender_id:string;recipient_id:string;content:string;created_at:string;read_at:string|null};

export default function MessagesPanel(){
 const [contacts,setContacts]=useState<Contact[]>([]);
 const [items,setItems]=useState<Message[]>([]);
 const [selected,setSelected]=useState("");
 const [draft,setDraft]=useState("");
 const [error,setError]=useState("");
 const [busy,setBusy]=useState(false);
 const [loading,setLoading]=useState(true);
 const fetchData=useCallback(async()=>{
  try{
   const [a,b]=await Promise.all([
    fetch("/api/messages/contacts",{cache:"no-store"}),
    fetch("/api/messages/inbox",{cache:"no-store"})
   ]);
   const [contactsPayload,messagesPayload]=await Promise.all([a.json(),b.json()]);
   if(!a.ok||!b.ok)throw new Error(contactsPayload.error||messagesPayload.error||"Không thể tải tin nhắn");
   const options=contactsPayload.items as Contact[];
   setContacts(options);setItems(messagesPayload.items as Message[]);
   setSelected(old=>options.some(c=>c.class_id+"|"+c.recipient_id===old)?old:options[0]?options[0].class_id+"|"+options[0].recipient_id:"");
   setError("");
  }catch(e){setError(e instanceof Error?e.message:"Không thể tải tin nhắn");}
  finally{setLoading(false);}
 },[]);
 useEffect(()=>{void fetchData();},[fetchData]);
 const active=contacts.find(c=>c.class_id+"|"+c.recipient_id===selected);
 const conversation=active?items.filter(m=>m.class_id===active.class_id&&(m.sender_id===active.recipient_id||m.recipient_id===active.recipient_id)).reverse():[];
 async function send(e:FormEvent<HTMLFormElement>){
  e.preventDefault();if(!active||!draft.trim()||busy)return;
  setBusy(true);setError("");
  try{
   const response=await fetch("/api/messages/inbox",{
    method:"POST",headers:{"Content-Type":"application/json"},
    body:JSON.stringify({class_id:active.class_id,recipient_id:active.recipient_id,content:draft.trim()})
   });
   const result=await response.json();
   if(!response.ok)throw new Error(result.error||"Không thể gửi");
   setDraft("");await fetchData();
  }catch(e){setError(e instanceof Error?e.message:"Không thể gửi");}
  finally{setBusy(false);}
 }
 async function markRead(id:string){
  setBusy(true);
  try{
   const response=await fetch("/api/messages/"+encodeURIComponent(id)+"/read",{method:"PATCH"});
   const result=await response.json();
   if(!response.ok)throw new Error(result.error||"Không thể xác nhận");
   await fetchData();
  }catch(e){setError(e instanceof Error?e.message:"Lỗi cập nhật");}
  finally{setBusy(false);}
 }
 return <Card className="messages-panel">
  <header className="messages-heading">
   <div><h2 className="icon-heading"><MessageCircle size={22} aria-hidden="true"/>Trao đổi nhà trường</h2>
    <p className="muted">Chỉ giáo viên phụ trách lớp và phụ huynh đã xác minh mới có thể trao đổi.</p>
   </div>
   <Button type="button" variant="outline" size="icon" onClick={()=>void fetchData()} aria-label="Làm mới tin nhắn"><RefreshCw size={18} aria-hidden="true"/></Button>
  </header>
  {error&&<p className="portal-message" role="alert">{error}</p>}
  {loading?<p role="status">Đang tải hội thoại…</p>:
   !contacts.length?<p className="muted">Chưa có liên hệ đủ điều kiện. Nhà trường cần phân công giáo viên và xác minh quan hệ phụ huynh – học sinh.</p>:
   <div className="messages-layout">
    <label className="messages-contacts">Người liên hệ
     <select value={selected} onChange={e=>setSelected(e.target.value)}>
      {contacts.map(c=><option key={c.class_id+"|"+c.recipient_id} value={c.class_id+"|"+c.recipient_id}>{c.name} · {c.class_name}</option>)}
     </select>
    </label>
    <div className="messages-thread" aria-live="polite">
     {conversation.length===0&&<p className="muted">Chưa có tin nhắn. Bạn có thể bắt đầu cuộc trao đổi.</p>}
     {conversation.map(msg=>{
      const mine=msg.recipient_id===active?.recipient_id;
      return <div className={"message-bubble "+(mine?"mine":"theirs")} key={msg.id}>
       <p>{msg.content}</p><small>{new Date(msg.created_at).toLocaleString("vi-VN")} · {mine?(msg.read_at?"Đã đọc":"Đã gửi"):(msg.read_at?"Đã đọc":"Chưa xác nhận")}</small>
       {!mine&&!msg.read_at&&<button type="button" disabled={busy} className="message-mark" onClick={()=>void markRead(msg.id)}><CheckCheck size={14} aria-hidden="true"/>Đánh dấu đã đọc</button>}
      </div>;
     })}
    </div>
    <form className="messages-compose" onSubmit={send}>
     <label htmlFor="message-content">Nội dung tin nhắn</label>
     <textarea id="message-content" required minLength={1} maxLength={2000} rows={3} value={draft} onChange={e=>setDraft(e.target.value)} placeholder="Nhập nội dung trao đổi phù hợp…" />
     <div className="messages-compose-actions"><small>{draft.length}/2000 ký tự</small><Button disabled={busy||!draft.trim()} type="submit"><Send size={17} aria-hidden="true"/>Gửi tin nhắn</Button></div>
    </form>
   </div>}
 </Card>;
}
