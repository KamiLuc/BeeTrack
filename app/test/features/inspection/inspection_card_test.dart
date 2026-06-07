import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/inspection/data/inspection_model.dart';
import 'package:app/features/inspection/view/inspection_history_screen.dart';
import 'package:app/l10n/app_localizations.dart';

class _MockApiClient extends Mock implements ApiClient {}

Inspection _insp({String? inspectedByName}) => Inspection(
      id: 1,
      hiveId: 1,
      inspectedAt: DateTime(2025, 6, 1),
      queenSeen: 'seen',
      broodPattern: '',
      aggressiveness: '',
      queenAdded: false,
      notes: '',
      inspectedByName: inspectedByName,
    );

Future<Widget> _wrap({required String myName, required Inspection inspection}) async {
  SharedPreferences.setMockInitialValues({'user_name': myName});
  final prefs = await SharedPreferences.getInstance();
  final storage = TokenStorage(prefs);
  return MultiRepositoryProvider(
    providers: [
      RepositoryProvider<TokenStorage>.value(value: storage),
      RepositoryProvider<ApiClient>.value(value: _MockApiClient()),
    ],
    child: MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: Scaffold(
        body: InspectionCard(
          apiaryId: 1,
          inspection: inspection,
          onEdit: () {},
          onDelete: () {},
        ),
      ),
    ),
  );
}

void main() {
  group('InspectionCard inspector name', () {
    testWidgets('shows "By <name>" when inspection was done by someone else', (tester) async {
      await tester.pumpWidget(await _wrap(
        myName: 'Alice',
        inspection: _insp(inspectedByName: 'Bob'),
      ));
      expect(find.textContaining('By Bob'), findsOneWidget);
    });

    testWidgets('does not show inspector name when it matches the current user', (tester) async {
      await tester.pumpWidget(await _wrap(
        myName: 'Alice',
        inspection: _insp(inspectedByName: 'Alice'),
      ));
      expect(find.textContaining('By Alice'), findsNothing);
    });

    testWidgets('does not show inspector name when it is null', (tester) async {
      await tester.pumpWidget(await _wrap(
        myName: 'Alice',
        inspection: _insp(inspectedByName: null),
      ));
      expect(find.textContaining('By'), findsNothing);
    });
  });
}
