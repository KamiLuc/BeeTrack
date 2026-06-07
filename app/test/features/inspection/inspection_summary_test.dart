import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:app/features/inspection/data/inspection_model.dart';
import 'package:app/features/inspection/view/inspection_summary.dart';
import 'package:app/l10n/app_localizations.dart';

Widget _wrap(Widget child) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: Scaffold(body: child),
    );

Inspection _insp({
  String queenSeen = '',
  String broodPattern = '',
  String aggressiveness = '',
  int? framesBrood,
  int? framesHoney,
  int? framesPollen,
  int? framesAddedDrawn,
  int? queenCellsCount,
  bool queenAdded = false,
  String notes = '',
}) =>
    Inspection(
      id: 1,
      hiveId: 1,
      inspectedAt: DateTime(2025, 6, 1),
      queenSeen: queenSeen,
      broodPattern: broodPattern,
      aggressiveness: aggressiveness,
      framesBrood: framesBrood,
      framesHoney: framesHoney,
      framesPollen: framesPollen,
      framesAddedDrawn: framesAddedDrawn,
      queenCellsCount: queenCellsCount,
      queenAdded: queenAdded,
      notes: notes,
    );

void main() {
  group('InspectionSummary', () {
    testWidgets('renders nothing when all fields are empty', (tester) async {
      await tester.pumpWidget(_wrap(InspectionSummary(inspection: _insp())));
      expect(find.byType(Text), findsNothing);
    });

    testWidgets('groups observations on one row with · separator', (tester) async {
      await tester.pumpWidget(_wrap(
        InspectionSummary(
          inspection: _insp(
            queenSeen: 'seen',
            broodPattern: 'good',
            aggressiveness: 'calm',
          ),
        ),
      ));
      expect(find.text('Queen seen · Brood: Medium · Calm'), findsOneWidget);
    });

    testWidgets('shows queen not seen status', (tester) async {
      await tester.pumpWidget(_wrap(
        InspectionSummary(inspection: _insp(queenSeen: 'not_seen')),
      ));
      expect(find.textContaining('Queen not seen'), findsOneWidget);
    });

    testWidgets('groups current frame counts on one row', (tester) async {
      await tester.pumpWidget(_wrap(
        InspectionSummary(
          inspection: _insp(framesBrood: 5, framesHoney: 3, framesPollen: 2),
        ),
      ));
      expect(
        find.text('Brood frames: 5 · Honey frames: 3 · Pollen frames: 2'),
        findsOneWidget,
      );
    });

    testWidgets('omits frame row when no frames set', (tester) async {
      await tester.pumpWidget(_wrap(
        InspectionSummary(inspection: _insp(queenSeen: 'seen')),
      ));
      expect(find.textContaining('frames:'), findsNothing);
    });

    testWidgets('shows queen added when true', (tester) async {
      await tester.pumpWidget(_wrap(
        InspectionSummary(inspection: _insp(queenAdded: true)),
      ));
      expect(find.text('Queen added'), findsOneWidget);
    });

    testWidgets('shows queen cells count', (tester) async {
      await tester.pumpWidget(_wrap(
        InspectionSummary(inspection: _insp(queenCellsCount: 4)),
      ));
      expect(find.textContaining('Queen cells: 4'), findsOneWidget);
    });

    testWidgets('shows note label and note text as separate widgets', (tester) async {
      await tester.pumpWidget(_wrap(
        InspectionSummary(inspection: _insp(notes: 'All good')),
      ));
      expect(find.text('Note'), findsOneWidget);
      expect(find.text('All good'), findsOneWidget);
    });

    testWidgets('shows date and time when showDate is true', (tester) async {
      await tester.pumpWidget(_wrap(
        InspectionSummary(
          inspection: _insp(queenSeen: 'seen'),
          showDate: true,
        ),
      ));
      expect(find.textContaining('Jun 1, 2025'), findsOneWidget);
      expect(find.textContaining('0:00'), findsOneWidget);
    });

    testWidgets('does not show date when showDate is false', (tester) async {
      await tester.pumpWidget(_wrap(
        InspectionSummary(
          inspection: _insp(queenSeen: 'seen'),
        ),
      ));
      expect(find.textContaining('2025'), findsNothing);
    });
  });
}
