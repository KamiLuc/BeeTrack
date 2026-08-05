import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:intl/intl.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/apiary/data/apiary_model.dart';
import 'package:app/features/hive/view/apiary_grid_screen.dart';
import 'package:app/l10n/app_localizations.dart';

class _RoutingAdapter implements HttpClientAdapter {
  _RoutingAdapter({
    required this.hivesJson,
    required List<Map<String, dynamic>> recordingsJson,
    this.inspectionJson,
  }) : recordingsJson = List.of(recordingsJson);

  final List<Map<String, dynamic>> hivesJson;
  final List<Map<String, dynamic>> recordingsJson;
  final Map<String, dynamic>? inspectionJson;
  final List<String> rejectedPaths = [];
  final List<String> acceptedPaths = [];
  final List<Map<String, dynamic>> updatedActionArguments = [];
  bool failAccept = false;
  bool failReject = false;
  bool failUpdateAction = false;

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    if (options.method == 'PATCH' && options.path.contains('/actions/')) {
      if (failUpdateAction) {
        return ResponseBody.fromString('{}', 500, headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        });
      }
      final segments = options.path.split('/');
      final actionId = int.parse(segments.last);
      final recordingId = int.parse(segments[segments.length - 3]);
      final data = options.data as Map<String, dynamic>;
      final toolArguments = data['tool_arguments'] as Map<String, dynamic>;
      updatedActionArguments.add(toolArguments);
      final recordingJson =
          recordingsJson.firstWhere((r) => r['recording_id'] == recordingId);
      final actionsJson =
          (recordingJson['voice_actions'] as List).cast<Map<String, dynamic>>();
      final actionJson =
          actionsJson.firstWhere((a) => a['id'] == actionId);
      final updatedActionJson = {...actionJson, 'tool_arguments': toolArguments};
      return ResponseBody.fromString(
        jsonEncode(updatedActionJson),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }
    if (options.path.endsWith('/reject')) {
      if (failReject) {
        return ResponseBody.fromString('{}', 500, headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        });
      }
      rejectedPaths.add(options.path);
      _removeResolvedRecording(options.path);
      return ResponseBody.fromString('{}', 200, headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      });
    }
    if (options.path.endsWith('/accept')) {
      if (failAccept) {
        return ResponseBody.fromString('{}', 500, headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        });
      }
      acceptedPaths.add(options.path);
      _removeResolvedRecording(options.path);
      return ResponseBody.fromString('{}', 200, headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      });
    }
    final Object body;
    if (options.path.endsWith('/voice-recordings')) {
      body = {'items': recordingsJson, 'total': recordingsJson.length};
    } else if (RegExp(r'/inspections/\d+$').hasMatch(options.path)) {
      body = inspectionJson!;
    } else if (RegExp(r'/hives/\d+$').hasMatch(options.path)) {
      body = hivesJson.first;
    } else {
      body = hivesJson;
    }
    return ResponseBody.fromString(
      jsonEncode(body),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  void _removeResolvedRecording(String path) {
    final segments = path.split('/');
    final recordingId = int.parse(segments[segments.length - 2]);
    recordingsJson.removeWhere((r) => r['recording_id'] == recordingId);
  }

  @override
  void close({bool force = false}) {}
}

Future<(ApiClient, _RoutingAdapter)> _fakeApiClient({
  required List<Map<String, dynamic>> hivesJson,
  required List<Map<String, dynamic>> recordingsJson,
  Map<String, dynamic>? inspectionJson,
}) async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  final apiClient = ApiClient(storage: TokenStorage(prefs), baseUrl: 'http://test');
  final adapter = _RoutingAdapter(
    hivesJson: hivesJson,
    recordingsJson: recordingsJson,
    inspectionJson: inspectionJson,
  );
  apiClient.dio.httpClientAdapter = adapter;
  return (apiClient, adapter);
}

Future<void> _solvePuzzleAndConfirm(WidgetTester tester, String confirmLabel) async {
  final promptFinder = find.byWidgetPredicate((w) =>
      w is Text && w.data != null && RegExp(r'^\d+ \+ \d+ = $').hasMatch(w.data!));
  final promptText = tester.widget<Text>(promptFinder).data!;
  final match = RegExp(r'(\d+) \+ (\d+)').firstMatch(promptText)!;
  final sum = int.parse(match.group(1)!) + int.parse(match.group(2)!);
  await tester.enterText(find.byType(TextField), '$sum');
  await tester.tap(find.widgetWithText(TextButton, confirmLabel).last);
  await tester.pumpAndSettle();
}

Widget _wrap(ApiClient apiClient, Widget child) => RepositoryProvider<ApiClient>.value(
      value: apiClient,
      child: MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('en'),
        home: child,
      ),
    );

const _apiary = Apiary(
  id: 1,
  name: 'Home apiary',
  lat: null,
  lng: null,
  gridRows: 2,
  gridCols: 2,
  hiveCount: 1,
  userRole: 'owner',
);

Map<String, dynamic> _hiveJson() => {
      'id': 1,
      'apiary_id': 1,
      'name': 'Alpha',
      'type': 'langstroth',
      'active': true,
      'grid_row': 0,
      'grid_col': 0,
    };

Map<String, dynamic> _completedRecordingJson() => {
      'recording_id': 7,
      'status': 'completed',
      'transcript': 'inspected hive alpha',
      'error_message': null,
      'created_at': '2026-08-01T12:30:00Z',
      'processed_at': '2026-08-01T12:31:00Z',
      'voice_actions': [
        {
          'id': 1,
          'sequence': 1,
          'hive_id': 1,
          'tool_name': 'create_inspection',
          'tool_arguments': {'colony_strength': 'strong', 'box_added': true},
          'status': 'proposed',
        },
      ],
    };

Map<String, dynamic> _noActionRecordingJson() => {
      'recording_id': 8,
      'status': 'completed',
      'transcript': 'just chatting, nothing to log',
      'error_message': null,
      'created_at': '2026-08-01T12:30:00Z',
      'processed_at': '2026-08-01T12:31:00Z',
      'voice_actions': [],
    };

Map<String, dynamic> _noArgumentsRecordingJson() => {
      'recording_id': 10,
      'status': 'completed',
      'transcript': 'marked hive as inactive',
      'error_message': null,
      'created_at': '2026-08-01T12:30:00Z',
      'processed_at': '2026-08-01T12:31:00Z',
      'voice_actions': [
        {
          'id': 3,
          'sequence': 1,
          'hive_id': 1,
          'tool_name': 'update_hive_status',
          'tool_arguments': null,
          'status': 'proposed',
        },
      ],
    };

Map<String, dynamic> _updateHiveStatusRecordingJson() => {
      'recording_id': 14,
      'status': 'completed',
      'transcript': 'hive is ready for harvest',
      'error_message': null,
      'created_at': '2026-08-01T12:30:00Z',
      'processed_at': '2026-08-01T12:31:00Z',
      'voice_actions': [
        {
          'id': 8,
          'sequence': 1,
          'hive_id': 1,
          'tool_name': 'update_hive_status',
          'tool_arguments': {'ready_for_harvest': true},
          'status': 'proposed',
        },
      ],
    };

Map<String, dynamic> _hiveNotIdentifiedRecordingJson() => {
      'recording_id': 9,
      'status': 'completed',
      'transcript': 'the bees looked fine',
      'error_message': null,
      'created_at': '2026-08-01T12:30:00Z',
      'processed_at': '2026-08-01T12:31:00Z',
      'voice_actions': [
        {
          'id': 2,
          'sequence': 1,
          'hive_id': null,
          'tool_name': null,
          'status': 'error',
          'error_message': 'HIVE_NOT_IDENTIFIED',
        },
      ],
    };

Map<String, dynamic> _proposalIncompleteRecordingJson() => {
      'recording_id': 11,
      'status': 'completed',
      'transcript': 'a very long rambling inspection note',
      'error_message': null,
      'created_at': '2026-08-01T12:30:00Z',
      'processed_at': '2026-08-01T12:31:00Z',
      'voice_actions': [
        {
          'id': 4,
          'sequence': 1,
          'hive_id': 1,
          'tool_name': null,
          'status': 'error',
          'error_message': 'PROPOSAL_INCOMPLETE',
        },
      ],
    };

Map<String, dynamic> _appliedInspectionRecordingJson() => {
      'recording_id': 12,
      'status': 'completed',
      'transcript': 'inspected hive alpha, already applied',
      'error_message': null,
      'created_at': '2026-08-01T12:30:00Z',
      'processed_at': '2026-08-01T12:31:00Z',
      'voice_actions': [
        {
          'id': 5,
          'sequence': 1,
          'hive_id': 1,
          'tool_name': 'create_inspection',
          'tool_arguments': {'colony_strength': 'strong'},
          'status': 'applied',
          'result_type': 'inspection',
          'result_record_id': 5,
        },
      ],
    };

Map<String, dynamic> _twoProposedActionsRecordingJson() => {
      'recording_id': 13,
      'status': 'completed',
      'transcript': 'inspected hive alpha and fed it sugar syrup',
      'error_message': null,
      'created_at': '2026-08-01T12:30:00Z',
      'processed_at': '2026-08-01T12:31:00Z',
      'voice_actions': [
        {
          'id': 6,
          'sequence': 1,
          'hive_id': 1,
          'tool_name': 'create_inspection',
          'tool_arguments': {'colony_strength': 'strong', 'box_added': true},
          'status': 'proposed',
        },
        {
          'id': 7,
          'sequence': 2,
          'hive_id': 1,
          'tool_name': 'create_feeding',
          'tool_arguments': {'feed_type': 'Sugar syrup (1:1)', 'amount': '2L'},
          'status': 'proposed',
        },
      ],
    };

Map<String, dynamic> _staleDiseasesInspectionRecordingJson() => {
      'recording_id': 14,
      'status': 'completed',
      'transcript': 'inspected hive alpha with stale diseases data',
      'error_message': null,
      'created_at': '2026-08-01T12:30:00Z',
      'processed_at': '2026-08-01T12:31:00Z',
      'voice_actions': [
        {
          'id': 8,
          'sequence': 1,
          'hive_id': 1,
          'tool_name': 'create_inspection',
          'tool_arguments': {
            'colony_strength': 'strong',
            'diseases': ['varroa'],
          },
          'status': 'proposed',
        },
      ],
    };

Map<String, dynamic> _inspectionJson() => {
      'id': 5,
      'hive_id': 1,
      'inspected_at': '2026-08-01T12:00:00Z',
      'colony_strength': 'strong',
      'notes': '',
    };

void main() {
  testWidgets(
      'shows a completed recording under "Ready for review" in the voice dialog',
      (tester) async {
    final (apiClient, _) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_completedRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    final expectedSubtitle = DateFormat('d.MM HH:mm')
        .format(DateTime.parse('2026-08-01T12:30:00Z').toLocal());

    expect(find.text('Ready for review'), findsWidgets);
    expect(find.text('inspected hive alpha'), findsOneWidget);
    expect(find.text(expectedSubtitle), findsOneWidget);
  });

  testWidgets('does not show the "Ready for review" section when there are no completed recordings',
      (tester) async {
    final (apiClient, _) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: const [],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    expect(find.text('Ready for review'), findsNothing);
  });

  testWidgets('shows "No action recognized" for a completed recording with no proposed actions',
      (tester) async {
    final (apiClient, _) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_noActionRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    expect(find.text('just chatting, nothing to log'), findsOneWidget);
    expect(find.text('No action recognized'), findsOneWidget);
  });

  testWidgets('shows a friendly error message for a HIVE_NOT_IDENTIFIED action',
      (tester) async {
    final (apiClient, _) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_hiveNotIdentifiedRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    expect(find.text('the bees looked fine'), findsOneWidget);
    expect(find.text("Couldn't identify the hive"), findsOneWidget);
  });

  testWidgets(
      'dismissing an error tile calls reject and removes it from the list',
      (tester) async {
    final (apiClient, adapter) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_hiveNotIdentifiedRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    expect(find.text('the bees looked fine'), findsOneWidget);

    await tester.tap(find.byTooltip('Dismiss'));
    await tester.pumpAndSettle();

    expect(adapter.rejectedPaths,
        contains('/api/v1/apiaries/1/voice-recordings/9/reject'));
    expect(find.text('the bees looked fine'), findsNothing);
  });

  testWidgets(
      'tapping a ready-for-review tile shows transcript and proposed action arguments',
      (tester) async {
    final (apiClient, _) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_completedRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('inspected hive alpha'));
    await tester.pumpAndSettle();

    expect(find.text('Recording details'), findsOneWidget);
    expect(find.text('Create inspection'), findsOneWidget);
    expect(find.text('Colony strength: Strong'), findsOneWidget);
    expect(find.text('Box added'), findsOneWidget);
  });

  testWidgets(
      'detail dialog content is width-capped even on a wide viewport',
      (tester) async {
    final originalSize = tester.view.physicalSize;
    final originalRatio = tester.view.devicePixelRatio;
    tester.view.physicalSize = const Size(1600, 1000);
    tester.view.devicePixelRatio = 1.0;
    addTearDown(() {
      tester.view.physicalSize = originalSize;
      tester.view.devicePixelRatio = originalRatio;
    });

    final (apiClient, _) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_completedRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('inspected hive alpha'));
    await tester.pumpAndSettle();

    final sizedBoxFinder = find.ancestor(
      of: find.text('Create inspection'),
      matching: find.byType(SizedBox),
    );
    final sizedBox = tester.widget<SizedBox>(sizedBoxFinder.first);
    expect(sizedBox.width, lessThan(1600));
    expect(sizedBox.width, 560.0);
  });

  testWidgets(
      'tapping a ready-for-review tile with no tool arguments shows the tool label with no argument lines',
      (tester) async {
    final (apiClient, _) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_noArgumentsRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('marked hive as inactive'));
    await tester.pumpAndSettle();

    expect(find.text('Recording details'), findsOneWidget);
    expect(find.text('Update hive status'), findsOneWidget);
    expect(find.text('Colony Strength: strong'), findsNothing);
  });

  testWidgets(
      'tapping a ready-for-review tile with an error action shows the friendly error message in the dialog',
      (tester) async {
    final (apiClient, _) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_hiveNotIdentifiedRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('the bees looked fine'));
    await tester.pumpAndSettle();

    expect(find.text('Recording details'), findsOneWidget);
    expect(find.text("Couldn't identify the hive"), findsNWidgets(2));
  });

  testWidgets(
      'accepting a normal-proposal recording from the detail dialog calls accept and removes it from the list',
      (tester) async {
    final (apiClient, adapter) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_completedRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('inspected hive alpha'));
    await tester.pumpAndSettle();

    expect(find.text('Accept'), findsOneWidget);
    await tester.tap(find.text('Accept'));
    await tester.pumpAndSettle();

    expect(adapter.acceptedPaths,
        contains('/api/v1/apiaries/1/voice-recordings/7/accept'));
    expect(find.text('Recording details'), findsNothing);
    expect(find.text('inspected hive alpha'), findsNothing);
  });

  testWidgets(
      'rejecting a normal-proposal recording from the detail dialog calls reject and removes it from the list',
      (tester) async {
    final (apiClient, adapter) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_completedRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('inspected hive alpha'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Reject'));
    await tester.pumpAndSettle();
    await _solvePuzzleAndConfirm(tester, 'Reject');

    expect(adapter.rejectedPaths,
        contains('/api/v1/apiaries/1/voice-recordings/7/reject'));
    expect(find.text('Recording details'), findsNothing);
    expect(find.text('inspected hive alpha'), findsNothing);
  });

  testWidgets(
      'rejecting a normal-proposal recording requires solving a math puzzle',
      (tester) async {
    final (apiClient, adapter) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_completedRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('inspected hive alpha'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Reject'));
    await tester.pumpAndSettle();

    expect(find.text('Reject recording?'), findsOneWidget);
    expect(adapter.rejectedPaths, isEmpty);

    // Enter a wrong answer first — should not reject.
    await tester.enterText(find.byType(TextField), '-1');
    await tester.tap(find.widgetWithText(TextButton, 'Reject').last);
    await tester.pumpAndSettle();
    expect(adapter.rejectedPaths, isEmpty);
    expect(find.text('Wrong answer'), findsOneWidget);
  });

  testWidgets(
      'dismissing an error/no-action recording does not require a math puzzle',
      (tester) async {
    final (apiClient, adapter) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_noActionRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('just chatting, nothing to log'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Reject'));
    await tester.pumpAndSettle();

    expect(find.text('Reject recording?'), findsNothing);
    expect(adapter.rejectedPaths,
        contains('/api/v1/apiaries/1/voice-recordings/8/reject'));
  });

  testWidgets(
      'a failed accept keeps the recording and shows an error toast',
      (tester) async {
    final (apiClient, adapter) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_completedRecordingJson()],
    );
    adapter.failAccept = true;

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('inspected hive alpha'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Accept'));
    await tester.pumpAndSettle();

    expect(find.text('Recording details'), findsOneWidget);
    expect(find.text('Failed to accept recording'), findsOneWidget);
  });

  testWidgets('no Accept button is shown for an error/no-action recording',
      (tester) async {
    final (apiClient, _) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_hiveNotIdentifiedRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('the bees looked fine'));
    await tester.pumpAndSettle();

    expect(find.text('Recording details'), findsOneWidget);
    expect(find.text('Accept'), findsNothing);
    expect(find.text('Reject'), findsOneWidget);
  });

  testWidgets(
      'the mic badge count drops immediately after accepting from the detail dialog, without closing the voice dialog',
      (tester) async {
    final (apiClient, _) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_completedRecordingJson(), _noActionRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    expect(find.text('2'), findsOneWidget);

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('inspected hive alpha'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Accept'));
    await tester.pumpAndSettle();

    expect(find.text('Recording details'), findsNothing);
    expect(find.text('Ready for review'), findsWidgets);
    expect(find.text('1'), findsOneWidget);
  });

  testWidgets(
      'a failed reject keeps the recording and shows an error toast',
      (tester) async {
    final (apiClient, adapter) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_completedRecordingJson()],
    );
    adapter.failReject = true;

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('inspected hive alpha'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Reject'));
    await tester.pumpAndSettle();
    await _solvePuzzleAndConfirm(tester, 'Reject');

    expect(find.text('Recording details'), findsOneWidget);
    expect(find.text('Failed to dismiss recording'), findsOneWidget);
  });

  testWidgets(
      'shows a friendly error message for a PROPOSAL_INCOMPLETE action',
      (tester) async {
    final (apiClient, _) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_proposalIncompleteRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('a very long rambling inspection note'));
    await tester.pumpAndSettle();

    expect(find.text('Recording details'), findsOneWidget);
    expect(
      find.textContaining('too long to process fully'),
      findsWidgets,
    );
  });

  testWidgets(
      'tapping an applied action with a result opens the underlying record for editing',
      (tester) async {
    final (apiClient, _) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_appliedInspectionRecordingJson()],
      inspectionJson: _inspectionJson(),
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('inspected hive alpha, already applied'));
    await tester.pumpAndSettle();

    expect(find.text('Create inspection'), findsOneWidget);
    expect(find.byIcon(Icons.chevron_right), findsOneWidget);

    await tester.tap(find.text('Create inspection'));
    await tester.pumpAndSettle();

    expect(find.text('Recording details'), findsNothing);
    expect(find.text('Edit inspection'), findsOneWidget);
  });

  testWidgets(
      'tapping a proposed create_inspection action opens the inspection form pre-filled, without closing the review dialog',
      (tester) async {
    final (apiClient, _) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_completedRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('inspected hive alpha'));
    await tester.pumpAndSettle();

    expect(find.byIcon(Icons.chevron_right), findsOneWidget);

    await tester.tap(find.text('Create inspection'));
    await tester.pumpAndSettle();

    expect(find.text('Edit inspection'), findsOneWidget);
    expect(find.widgetWithText(TextFormField, 'Notes'), findsOneWidget);

    await tester.pageBack();
    await tester.pumpAndSettle();

    expect(find.text('Recording details'), findsOneWidget);
    expect(find.text('Create inspection'), findsOneWidget);
  });

  testWidgets(
      'does not render a diseases line for a create_inspection card even '
      'if tool_arguments has a stale "diseases" key from before it was '
      'removed from the schema', (tester) async {
    final (apiClient, _) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_staleDiseasesInspectionRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('inspected hive alpha with stale diseases data'));
    await tester.pumpAndSettle();

    expect(find.text('Recording details'), findsOneWidget);
    expect(find.text('Diseases'), findsNothing);
  });

  testWidgets(
      'tapping one proposed action card shows a spinner on that card only, leaving other cards visible',
      (tester) async {
    final (apiClient, _) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_twoProposedActionsRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('inspected hive alpha and fed it sugar syrup'));
    await tester.pumpAndSettle();

    expect(find.text('Create inspection'), findsOneWidget);
    expect(find.text('Create feeding'), findsOneWidget);
    expect(find.text('Sugar syrup (1:1) · 2L'), findsOneWidget);

    await tester.tap(find.text('Create inspection'));
    await tester.pump();

    expect(find.byType(CircularProgressIndicator), findsOneWidget);
    expect(find.text('Create feeding'), findsOneWidget);
    expect(find.text('Sugar syrup (1:1) · 2L'), findsOneWidget);

    await tester.pumpAndSettle();
  });

  testWidgets(
      'saving an edited proposed inspection updates the action arguments and refreshes the card in place',
      (tester) async {
    final (apiClient, adapter) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_completedRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('inspected hive alpha'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Create inspection'));
    await tester.pumpAndSettle();

    await tester.enterText(
      find.widgetWithText(TextFormField, 'Notes'),
      'queen looks great',
    );
    await tester.pump();

    await tester.tap(find.byIcon(Icons.check));
    await tester.pumpAndSettle();

    expect(find.text('Recording details'), findsOneWidget);
    expect(adapter.updatedActionArguments, hasLength(1));
    expect(adapter.updatedActionArguments.single['notes'], 'queen looks great');
    expect(find.text('queen looks great'), findsOneWidget);
  });

  testWidgets(
      'tapping a proposed update_hive_status action opens the hive status form pre-filled, without closing the review dialog',
      (tester) async {
    final (apiClient, _) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_updateHiveStatusRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('hive is ready for harvest'));
    await tester.pumpAndSettle();

    expect(find.byIcon(Icons.chevron_right), findsOneWidget);

    await tester.tap(find.text('Update hive status'));
    await tester.pumpAndSettle();

    expect(find.text('Recording details'), findsNothing);
    expect(find.text('Ready for harvest'), findsOneWidget);

    await tester.pageBack();
    await tester.pumpAndSettle();

    expect(find.text('Recording details'), findsOneWidget);
    expect(find.text('Update hive status'), findsOneWidget);
  });

  testWidgets(
      'saving an edited proposed update_hive_status action updates the card in place',
      (tester) async {
    final (apiClient, adapter) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_updateHiveStatusRecordingJson()],
    );

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('hive is ready for harvest'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Update hive status'));
    await tester.pumpAndSettle();

    final needsFoodSwitch = find.descendant(
      of: find.ancestor(
        of: find.text('Needs food'),
        matching: find.byType(Row),
      ).first,
      matching: find.byType(Switch),
    );
    await tester.tap(needsFoodSwitch);
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.check));
    await tester.pumpAndSettle();

    expect(find.text('Recording details'), findsOneWidget);
    expect(adapter.updatedActionArguments, hasLength(1));
    expect(adapter.updatedActionArguments.single['ready_for_harvest'], true);
    expect(adapter.updatedActionArguments.single['needs_food'], true);
  });

  testWidgets(
      'a failed update while editing a proposed action shows the form error and leaves it open',
      (tester) async {
    final (apiClient, adapter) = await _fakeApiClient(
      hivesJson: [_hiveJson()],
      recordingsJson: [_completedRecordingJson()],
    );
    adapter.failUpdateAction = true;

    await tester.pumpWidget(_wrap(apiClient, const ApiaryGridScreen(apiary: _apiary)));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.mic_none));
    await tester.pumpAndSettle();

    await tester.tap(find.text('inspected hive alpha'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Create inspection'));
    await tester.pumpAndSettle();

    await tester.tap(find.byIcon(Icons.check));
    await tester.pumpAndSettle();

    expect(find.text('Edit inspection'), findsOneWidget);
    expect(find.byType(SnackBar), findsOneWidget);
  });
}
