# OpenEdu Flutter mobile foundation (0.4)

This is an **iOS Simulator-built, widget-tested native mobile foundation**, not an App Store or Google Play release. Both platforms share one Dart codebase with an iOS-inspired frosted/glass design and the Lucide icon family (`lucide_icons_flutter`).

## Implemented
- Public permitted-book search and chapter/lesson browsing
- School login with optional explicit UUID for multi-school users
- Token stored using `flutter_secure_storage` (iOS Keychain, Android Keystore plugin), never shared preferences
- School-scoped classes and targeted announcements
- Verified guardian/assigned teacher messaging, contact selection and sending
- Device-local logout and JWT session checking after relaunch
- Home navigation for library, school, messaging and account, with empty/offline states

## Native launch configuration
The API has **no public endpoint configured for mobile**. Do not copy a private Railway `.internal` URL into a device. Work with the operator to configure a reviewed HTTPS API domain, apply API rate limits and secret policies, and verify authorization before real user data is used.

```sh
cd apps/mobile
flutter pub get
flutter test
flutter build ios --simulator
flutter run -d <ios-simulator-identifier> --dart-define=OPENEDU_API_URL=https://approved-api.example.org
```

Never embed passwords, JWT secrets, private Railway URLs, database credentials or production auth tokens into the app. API URL is public configuration only. Production URLs require HTTPS. Private API remains internal until explicitly secured/exposed.

## Requirements and limitations
- This code has **no offline textbook cache**, push notifications, biometric login, background sync, quizzes, files/PDF reader, complete homework workflow or admin dashboard.
- Native refresh tokens are not implemented; API JWTs expire after 15 minutes. Session expiry requires reauthentication.
- Flutter tests and unsigned iOS Simulator build pass on Mac. Android SDK is **missing** on the current development host; Android build and Play Store acceptance are unverified.
- `flutter analyze` encountered a Dart analysis-server process error on this host despite successful widget tests and Xcode Simulator build. Investigate Flutter SDK state or run static analysis on CI with a clean SDK before merging to main.
- Compliance for minors, guardian verification, app privacy disclosures and production onboarding are required before storing real student records.

## Personal lesson bookmarks and progress

Open a published text lesson from the library, then use `Lưu bài` or `Đánh dấu đã học` in the native lesson dialog. Both flags are sent to the Go API over authenticated HTTPS; they are not persisted in plaintext on the phone. Guests can still read *public* permitted text but cannot save personal progress. School-only lessons require an active logged-in session, independent of the saved state.

## Timed school exams

Open a class in the School tab to see its published timed quizzes. Enrolled students can start a server-timed attempt, choose one of four answers per question, submit once, and view the server-calculated score. Correct answers are never included in student API responses. Exam creation/editing and staff results remain on the Web dashboard until complete native staff UX is implemented. API endpoints are in `docs/QUIZ_EXAMS.md`. Flutter widget tests validate app navigation; native quiz device acceptance requires a configured HTTPS mobile API endpoint.
