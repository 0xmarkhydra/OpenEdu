import 'dart:ui';
import 'package:flutter/material.dart';
import 'package:lucide_icons_flutter/lucide_icons.dart';
import 'core/api.dart';
import 'features/reading/lesson_dialog.dart';
import 'features/quizzes/quiz_screen.dart';

const ink = Color(0xFF1A2E54);
const blue = Color(0xFF4B80EF);
const violet = Color(0xFF8066E9);

void main() => runApp(const OpenEduApp());

class OpenEduApp extends StatelessWidget {
  const OpenEduApp({super.key});
  @override
  Widget build(BuildContext context) => MaterialApp(
    title: 'OpenEdu',
    debugShowCheckedModeBanner: false,
    theme: ThemeData(
      colorScheme: ColorScheme.fromSeed(seedColor: blue),
      fontFamily: '.SF Pro Text',
      useMaterial3: true,
      scaffoldBackgroundColor: const Color(0xFFF3F7FF),
      inputDecorationTheme: InputDecorationTheme(
        filled: true,
        fillColor: Colors.white.withValues(alpha: .78),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(15)),
        contentPadding: const EdgeInsets.symmetric(horizontal: 15, vertical: 13),
      ),
    ),
    home: const OpenEduHome(),
  );
}

class Frosted extends StatelessWidget {
  final Widget child;
  final EdgeInsets padding;
  const Frosted({super.key, required this.child, this.padding = const EdgeInsets.all(20)});
  @override
  Widget build(BuildContext context) => ClipRRect(
    borderRadius: BorderRadius.circular(25),
    child: BackdropFilter(
      filter: ImageFilter.blur(sigmaX: 18, sigmaY: 18),
      child: DecoratedBox(
        decoration: BoxDecoration(
          borderRadius: BorderRadius.circular(25),
          color: Colors.white.withValues(alpha: .69),
          border: Border.all(color: Colors.white.withValues(alpha: .92)),
          boxShadow: [BoxShadow(color: blue.withValues(alpha: .08),blurRadius: 26,offset: const Offset(0,12))],
        ),
        child: Padding(padding: padding,child: child),
      ),
    ),
  );
}

class OpenEduHome extends StatefulWidget {
  const OpenEduHome({super.key});
  @override
  State<OpenEduHome> createState() => _OpenEduHomeState();
}
class _OpenEduHomeState extends State<OpenEduHome> {
  final api = OpenEduApi();
  final search = TextEditingController();
  final email = TextEditingController();
  final password = TextEditingController();
  final schoolId = TextEditingController();
  final draft = TextEditingController();
  int tab = 0;
  bool loading = true;
  bool busy = false;
  String? error;
  String? selectedContact;
  List<Map<String,dynamic>> books=[];
  List<Map<String,dynamic>> classes=[];
  List<Map<String,dynamic>> announcements=[];
  List<Map<String,dynamic>> messages=[];
  List<Map<String,dynamic>> contacts=[];

  @override
  void initState(){
    super.initState();
    _bootstrap();
  }
  Future<void> _bootstrap() async {
    try { await api.restoreSession(); } catch (_) { /* Offline must not crash. */ }
    await refresh();
  }
  @override
  void dispose(){
    api.dispose();search.dispose();email.dispose();password.dispose();schoolId.dispose();draft.dispose();
    super.dispose();
  }
  Future<void> refresh()async{
    if(mounted)setState(()=>loading=true);
    try{
      final found=await api.books(query:search.text);
      List<Map<String,dynamic>> cls=[],news=[],msgs=[],allowed=[];
      if(api.signedIn){
        if(api.schoolId?.isNotEmpty == true){
          cls=await api.schoolClasses();
          news=await api.announcements();
        }
        if(api.roles.any((r)=>r=='guardian'||r=='teacher')){
          msgs=await api.messages();
          final response=await api.request('GET','/v1/me/messages/contacts');
          allowed=(response['items'] as List<dynamic>? ?? [])
            .map((x)=>Map<String,dynamic>.from(x as Map)).toList();
        }
      }
      if(mounted)setState((){
        books=found;classes=cls;announcements=news;messages=msgs;contacts=allowed;
        selectedContact ??=allowed.isNotEmpty?'0':null;error=null;
      });
    }catch(e){if(mounted)setState(()=>error=e.toString());}
    finally{if(mounted)setState(()=>loading=false);}
  }
  Future<void> signIn()async{
    setState(()=>busy=true);
    try{
      await api.login(email.text,password.text,schoolId:schoolId.text);
      password.clear();
      if(mounted)setState(()=>tab=1);
      await refresh();
    }catch(e){if(mounted)setState(()=>error=e.toString());}
    finally{if(mounted)setState(()=>busy=false);}
  }
  Future<void> logOut()async{
    await api.signOut();
    if(mounted)setState((){tab=0;classes=[];announcements=[];messages=[];contacts=[];selectedContact=null;});
    await refresh();
  }
  Future<void> send()async{
    final i=int.tryParse(selectedContact??'');
    if(i==null||i<0||i>=contacts.length||draft.text.trim().isEmpty)return;
    setState(()=>busy=true);
    try{
      final contact=contacts[i];
      await api.request('POST','/v1/me/messages',body:{
        'class_id':contact['class_id'],
        'recipient_id':contact['recipient_id'],
        'content':draft.text.trim(),
      });
      draft.clear();await refresh();
    }catch(e){if(mounted)setState(()=>error=e.toString());}
    finally{if(mounted)setState(()=>busy=false);}
  }
  Widget empty(String content)=>Frosted(child:Row(children:[
    const Icon(LucideIcons.info,color:blue),const SizedBox(width:12),
    Expanded(child:Text(content,style:const TextStyle(color:ink,height:1.55))),
  ]));
  Widget header(){
    final title=['Thư viện mở','Không gian học tập','Trao đổi','Tài khoản'][tab];
    return Column(crossAxisAlignment:CrossAxisAlignment.start,children:[
      const SizedBox(height:18),
      Row(children:[
        Container(width:43,height:43,decoration:BoxDecoration(borderRadius:BorderRadius.circular(15),
          gradient:const LinearGradient(colors:[Color(0xFF5CC2F8),blue,violet])),
          child:const Icon(LucideIcons.bookOpenCheck,color:Colors.white,size:25)),
        const SizedBox(width:11),
        const Text('OpenEdu',style:TextStyle(fontSize:23,fontWeight:FontWeight.w800,color:ink,letterSpacing:-.8)),
        const Spacer(),
        IconButton(tooltip:'Làm mới',onPressed:loading?null:refresh,
          icon:const Icon(LucideIcons.refreshCw,color:blue)),
      ]),
      const SizedBox(height:32),
      Text(title,style:const TextStyle(fontSize:34,fontWeight:FontWeight.w800,letterSpacing:-1.2,color:ink)),
      const SizedBox(height:8),
      Text(['Khám phá học liệu có quyền phân phối',
        'Lớp học, bài tập và thông báo từ nhà trường',
        'Kết nối giáo viên và phụ huynh đã xác minh',
        'Bảo mật và đăng nhập OpenEdu'][tab],
        style:const TextStyle(fontSize:14,color:Color(0xFF6B7C99),height:1.5)),
      const SizedBox(height:25),
    ]);
  }
  Widget library(){
    return Column(children:[
      Frosted(child:Column(children:[
        TextField(controller:search,textInputAction:TextInputAction.search,onSubmitted:(_)=>refresh(),
          decoration:const InputDecoration(labelText:'Tên sách hoặc bài học',
            prefixIcon:Icon(LucideIcons.search))),
        const SizedBox(height:12),
        SizedBox(width:double.infinity,child:FilledButton.icon(onPressed:refresh,
          icon:const Icon(LucideIcons.search,size:18),label:const Text('Tìm học liệu'))),
      ])),
      const SizedBox(height:17),
      if(books.isEmpty)empty('Chưa có học liệu phù hợp hoặc chưa kết nối được thư viện. OpenEdu chỉ hiển thị nội dung đã được cấp phép.')
      else ...books.map((book)=>Padding(
        padding:const EdgeInsets.only(bottom:12),
        child:Frosted(child:ListTile(contentPadding:EdgeInsets.zero,
          leading:const Icon(LucideIcons.bookOpen,size:30,color:blue),
          title:Text(book['title']?.toString()??'',style:const TextStyle(fontWeight:FontWeight.w700)),
          subtitle:Text('Lớp ${book['grade']} · ${book['subject_code']}'),
          trailing:const Icon(LucideIcons.chevronRight,color:blue),
          onTap:()=>showBook(book),
        )),
      )),
    ]);
  }
  Future<void> showBook(Map<String,dynamic> book) async{
    final id=book['id']?.toString();
    if(id==null)return;
    try{
      final outline=await api.request('GET','/v1/books/$id/outline');
      if(!mounted)return;
      final chapters=outline['items'] as List<dynamic>? ?? [];
      await showModalBottomSheet<void>(context:context,isScrollControlled:true,backgroundColor:const Color(0xFFF5F8FF),
       shape:const RoundedRectangleBorder(borderRadius:BorderRadius.vertical(top:Radius.circular(26))),
       builder:(sheet)=>SafeArea(child:Padding(padding:const EdgeInsets.all(22),child:ListView(
        shrinkWrap:true,children:[
          Text(book['title']?.toString()??'Bài học',
            style:const TextStyle(fontSize:23,fontWeight:FontWeight.w800,color:ink)),
          const SizedBox(height:15),
          if(chapters.isEmpty)const Text('Chưa phát hành nội dung chương/bài.'),
          for(final chapter in chapters) ...[
            Padding(padding:const EdgeInsets.symmetric(vertical:13),
              child:Text(chapter['title']?.toString()??'',style:const TextStyle(fontWeight:FontWeight.w700))),
            for(final lesson in (chapter['lessons'] as List<dynamic>? ?? []))
              ListTile(title:Text(lesson['title']?.toString()??''),
                leading:const Icon(LucideIcons.bookOpen,size:19,color:blue),
                onTap:() async{
                  try{
                    final content=await api.request('GET','/v1/lessons/${lesson['id']}');
                    if(!sheet.mounted)return;
                    await showDialog<void>(context:sheet,builder:(dialog)=>LessonDialog(
                      api:api,bookId:id,lessonId:lesson['id'].toString(),
                      title:content['title']?.toString()??'Bài học',
                      content:content['content']?.toString()??''));
                  }catch(e){if(sheet.mounted)ScaffoldMessenger.of(sheet).showSnackBar(SnackBar(content:Text(e.toString())));}
                }),
          ],
        ]))));
    }catch(e){if(mounted)setState(()=>error=e.toString());}
  }
  Widget school(){
    if(!api.signedIn)return empty('Đăng nhập ở tab Tài khoản để xem thông tin của trường.');
    if(api.schoolId?.isNotEmpty!=true)return empty('Tài khoản chưa được cấp quyền trong trường.');
    return Column(crossAxisAlignment:CrossAxisAlignment.start,children:[
      const Text('Lớp học',style:TextStyle(fontSize:20,fontWeight:FontWeight.w800,color:ink)),
      const SizedBox(height:12),
      if(classes.isEmpty)empty('Bạn chưa có lớp học hoặc đang chờ được phân công.')
      else ...classes.map((c)=>Padding(padding:const EdgeInsets.only(bottom:10),
        child:Frosted(child:ListTile(
          contentPadding:EdgeInsets.zero,
          leading:const Icon(LucideIcons.school,color:blue),
          title:Text(c['name']?.toString()??'Lớp học',style:const TextStyle(fontWeight:FontWeight.w700)),
          subtitle:Text('Khối ${c['grade']} · ${c['academic_year']}'),
          trailing:const Icon(LucideIcons.chevronRight,color:blue),
          onTap:()=>Navigator.push<void>(context,MaterialPageRoute(
            builder:(_)=>QuizScreen(api:api,classId:c['id'].toString(),
              className:c['name']?.toString()??'Lớp học'))),
        )))),
      const SizedBox(height:25),
      const Text('Thông báo',style:TextStyle(fontSize:20,fontWeight:FontWeight.w800,color:ink)),
      const SizedBox(height:12),
      if(announcements.isEmpty)empty('Nhà trường chưa có thông báo.')
      else ...announcements.map((n)=>Padding(padding:const EdgeInsets.only(bottom:10),
        child:Frosted(child:Column(crossAxisAlignment:CrossAxisAlignment.start,children:[
          Row(children:[const Icon(LucideIcons.bell,color:blue,size:18),const SizedBox(width:8),
            Expanded(child:Text(n['title']?.toString()??'',style:const TextStyle(fontWeight:FontWeight.w700)))]),
          const SizedBox(height:10),Text(n['content']?.toString()??'',style:const TextStyle(height:1.5)),
        ])))),
    ]);
  }
  Widget messagesTab(){
    if(!api.signedIn)return empty('Đăng nhập để xem tin nhắn.');
    if(!api.roles.any((role)=>role=='guardian'||role=='teacher'))return empty('Chức năng này dành cho giáo viên hoặc phụ huynh có quan hệ giám hộ đã xác minh.');
    return Column(children:[
      if(contacts.isEmpty)empty('Chưa có liên hệ được xác minh. Nhà trường cần liên kết phụ huynh và phân công giáo viên.')
      else Frosted(child:Column(crossAxisAlignment:CrossAxisAlignment.stretch,children:[
        DropdownButtonFormField<String>(
          initialValue:selectedContact,
          decoration:const InputDecoration(labelText:'Người liên hệ'),
          items:[for(var i=0;i<contacts.length;i++)DropdownMenuItem(
            value:'$i',child:Text('${contacts[i]['name']} · ${contacts[i]['class_name']}',
              overflow:TextOverflow.ellipsis))],
          onChanged:(value)=>setState(()=>selectedContact=value),
        ),
        const SizedBox(height:12),
        TextField(controller:draft,maxLines:3,maxLength:2000,
          decoration:const InputDecoration(labelText:'Nội dung trao đổi')),
        FilledButton.icon(onPressed:busy?null:send,icon:const Icon(LucideIcons.send,size:18),
          label:const Text('Gửi tin nhắn')),
      ])),
      const SizedBox(height:16),
      ...messages.take(50).map((m)=>Padding(padding:const EdgeInsets.only(bottom:10),
        child:Frosted(child:Column(crossAxisAlignment:CrossAxisAlignment.start,children:[
          Text(m['sender_id']==api.userId?'Bạn gửi':'Tin nhắn nhận',
            style:const TextStyle(color:blue,fontWeight:FontWeight.w700)),
          const SizedBox(height:6),Text(m['content']?.toString()??'',style:const TextStyle(height:1.55)),
          const SizedBox(height:6),Text(m['created_at']?.toString()??'',
            style:const TextStyle(color:Color(0xFF8292A8),fontSize:11)),
        ])))),
    ]);
  }
  Widget profile(){
    if(!api.configured)return empty('API mobile chưa cấu hình. Cần URL API HTTPS được bảo vệ trước khi thử đăng nhập trên thiết bị thật.');
    if(api.signedIn)return Frosted(child:Column(crossAxisAlignment:CrossAxisAlignment.stretch,children:[
      const Icon(LucideIcons.shieldCheck,size:43,color:blue),
      const SizedBox(height:12),
      const Text('Đã đăng nhập',textAlign:TextAlign.center,style:TextStyle(fontWeight:FontWeight.w800,fontSize:22,color:ink)),
      const SizedBox(height:12),
      Text('Vai trò: ${api.roles.join(", ")}',textAlign:TextAlign.center),
      const SizedBox(height:10),
      FilledButton.tonalIcon(onPressed:logOut,icon:const Icon(LucideIcons.logOut),
        label:const Text('Đăng xuất an toàn')),
    ]));
    return Frosted(child:Form(child:Column(crossAxisAlignment:CrossAxisAlignment.stretch,children:[
      const Text('Đăng nhập OpenEdu',style:TextStyle(fontSize:22,fontWeight:FontWeight.w800,color:ink)),
      const SizedBox(height:20),
      TextField(controller:email,keyboardType:TextInputType.emailAddress,autocorrect:false,
        decoration:const InputDecoration(labelText:'Email')),
      const SizedBox(height:12),
      TextField(controller:password,obscureText:true,autocorrect:false,
        decoration:const InputDecoration(labelText:'Mật khẩu')),
      const SizedBox(height:12),
      TextField(controller:schoolId,autocorrect:false,
        decoration:const InputDecoration(labelText:'Mã trường UUID (nếu có nhiều trường)')),
      const SizedBox(height:20),
      FilledButton.icon(onPressed:busy?null:signIn,icon:const Icon(LucideIcons.logIn,size:18),
        label:Text(busy?'Đang đăng nhập…':'Đăng nhập')),
      const SizedBox(height:9),
      const Text('Token chỉ được lưu trong Keychain hoặc kho lưu trữ bảo mật của thiết bị.',
        style:TextStyle(color:Color(0xFF6F8099),fontSize:12,height:1.5)),
    ])));
  }
  @override
  Widget build(BuildContext context){
    return Scaffold(
      body:Stack(children:[
        const Positioned.fill(child:DecoratedBox(decoration:BoxDecoration(
          gradient:LinearGradient(begin:Alignment.topLeft,end:Alignment.bottomRight,
            colors:[Color(0xFFE5F2FF),Color(0xFFF9F8FF),Color(0xFFEDEFFF)])))),
        SafeArea(child:RefreshIndicator(onRefresh:refresh,
          child:ListView(padding:const EdgeInsets.symmetric(horizontal:18,vertical:9),children:[
            header(),
            if(error!=null)Padding(padding:const EdgeInsets.only(bottom:14),
              child:Frosted(child:Text(error!,style:const TextStyle(color:Color(0xFFAF3855))))),
            if(loading)const LinearProgressIndicator(minHeight:2),
            const SizedBox(height:12),
            [library,school,messagesTab,profile][tab](),
            const SizedBox(height:38),
          ]))),
      ]),
      bottomNavigationBar:SafeArea(top:false,child:NavigationBar(
        backgroundColor:const Color(0xFFF7F9FF),
        indicatorColor:const Color(0xFFDFE9FF),
        selectedIndex:tab,onDestinationSelected:(v)=>setState(()=>tab=v),
        destinations:const [
          NavigationDestination(icon:Icon(LucideIcons.libraryBig),label:'Thư viện'),
          NavigationDestination(icon:Icon(LucideIcons.school),label:'Lớp học'),
          NavigationDestination(icon:Icon(LucideIcons.messageCircle),label:'Tin nhắn'),
          NavigationDestination(icon:Icon(LucideIcons.userRound),label:'Tài khoản'),
        ],
      )),
    );
  }
}