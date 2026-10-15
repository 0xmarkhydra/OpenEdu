import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:openedu/main.dart';

void main() {
  testWidgets('OpenEdu renders branded mobile navigation', (tester) async {
    await tester.pumpWidget(const OpenEduApp());
    await tester.pump();
    expect(find.text('OpenEdu'), findsOneWidget);
    expect(find.text('Thư viện mở'), findsOneWidget);
    expect(find.byType(NavigationBar), findsOneWidget);
    expect(find.text('Thư viện'), findsOneWidget);
    expect(find.text('Tài khoản'), findsOneWidget);
  });

  testWidgets('OpenEdu switches to account screen', (tester) async {
    await tester.pumpWidget(const OpenEduApp());
    await tester.pump();
    await tester.tap(find.text('Tài khoản'));
    await tester.pump();
    expect(find.text('Bảo mật và đăng nhập OpenEdu'), findsOneWidget);
  });
}
