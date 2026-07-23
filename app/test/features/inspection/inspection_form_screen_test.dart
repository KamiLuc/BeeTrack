import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:image_picker/image_picker.dart';
import 'package:image_picker_platform_interface/image_picker_platform_interface.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/hive/data/hive_model.dart';
import 'package:app/features/inspection/data/inspection_model.dart';
import 'package:app/features/inspection/view/inspection_form_screen.dart';
import 'package:app/l10n/app_localizations.dart';

/// A valid 1x1 transparent PNG, so `Image.memory` can decode the picked
/// files' bytes in the pending thumbnails without throwing.
final Uint8List _kOnePixelPng = base64Decode(
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY'
  '42YAAAAASUVORK5CYII=',
);

/// A fake image_picker platform that returns a new in-memory [XFile] on
/// every pick, so photo-picker flows can be exercised without a real
/// device/plugin.
class _FakeImagePickerPlatform extends ImagePickerPlatform {
  int pickCount = 0;

  @override
  Future<XFile?> getImageFromSource({
    required ImageSource source,
    ImagePickerOptions options = const ImagePickerOptions(),
  }) async {
    pickCount++;
    return XFile.fromData(
      _kOnePixelPng,
      name: 'photo$pickCount.jpg',
      mimeType: 'image/jpeg',
    );
  }
}

class _RecordingAdapter implements HttpClientAdapter {
  final List<RequestOptions> requests = [];
  List<Map<String, dynamic>> images = [];

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);

    if (options.path.contains('/invitations/count')) {
      return ResponseBody.fromString(
        jsonEncode({'count': 0}),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }

    if (options.path.contains('/inspections') && options.method == 'POST') {
      return ResponseBody.fromString(
        jsonEncode({
          'id': 99,
          'hive_id': 10,
          'inspected_at': DateTime(2025, 6, 1).toUtc().toIso8601String(),
          'queen_status': 'seen',
          'brood_pattern': '',
          'aggressiveness': '',
          'queen_added': false,
          'notes': '',
        }),
        201,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }

    if (options.path.contains('/images') && options.method == 'GET') {
      return ResponseBody.fromString(
        jsonEncode(images),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }

    if (options.path.contains('/images') && options.method == 'POST') {
      return ResponseBody.fromString(
        jsonEncode({
          'id': 7,
          'inspection_id': 99,
          'mime_type': 'image/jpeg',
          'created_at': DateTime(2026, 1, 1).toIso8601String(),
        }),
        201,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }

    // Generic success for PATCH hive update, image delete, etc.
    return ResponseBody.fromString(
      jsonEncode({}),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

Future<(ApiClient, _RecordingAdapter)> _fakeApiClient() async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  final apiClient =
      ApiClient(storage: TokenStorage(prefs), baseUrl: 'http://test');
  final adapter = _RecordingAdapter();
  apiClient.dio.httpClientAdapter = adapter;
  return (apiClient, adapter);
}

Widget _wrap(ApiClient apiClient, Widget child) => MaterialApp(
      localizationsDelegates: AppLocalizations.localizationsDelegates,
      supportedLocales: AppLocalizations.supportedLocales,
      locale: const Locale('en'),
      home: RepositoryProvider<ApiClient>.value(
        value: apiClient,
        child: child,
      ),
    );

const _hive = Hive(
  id: 10,
  apiaryId: 1,
  name: 'Alpha',
  type: 'langstroth',
  active: true,
  queenless: false,
  readyForHarvest: false,
  needsFood: false,
  gridRow: 0,
  gridCol: 0,
);

// Index order of the +/- IconButton pairs rendered by `_SignedFrameField`
// in the frames section (drawn, foundation, brood, feed).
const _drawnField = 0;
const _foundationField = 1;

Future<void> _tapAdd(WidgetTester tester, int fieldIndex, {int times = 1}) async {
  final finder = find.byIcon(Icons.add).at(fieldIndex);
  for (var i = 0; i < times; i++) {
    await tester.ensureVisible(finder);
    await tester.tap(finder);
    await tester.pump();
  }
}

Future<void> _tapRemove(WidgetTester tester, int fieldIndex, {int times = 1}) async {
  final finder = find.byIcon(Icons.remove).at(fieldIndex);
  for (var i = 0; i < times; i++) {
    await tester.ensureVisible(finder);
    await tester.tap(finder);
    await tester.pump();
  }
}

void main() {
  late AppLocalizations l10n;
  setUpAll(() async {
    l10n = await AppLocalizations.delegate.load(const Locale('en'));
  });

  group('InspectionFormScreen signed frame delta logic', () {
    testWidgets(
        'tapping minus decrements the field and flips the label to removed',
        (tester) async {
      final (apiClient, _) = await _fakeApiClient();

      await tester.pumpWidget(_wrap(
        apiClient,
        const InspectionFormScreen(apiaryId: 1, hive: _hive),
      ));
      await tester.pumpAndSettle();

      expect(find.text('Added empty frames'), findsOneWidget);

      await _tapRemove(tester, _drawnField);
      expect(find.text('Taken empty frames'), findsOneWidget);
      expect(find.text('-1'), findsOneWidget);

      await _tapAdd(tester, _drawnField);
      expect(find.text('Added empty frames'), findsOneWidget);
      expect(find.text('0'), findsWidgets);
    });

    testWidgets(
        'tapping plus increments the displayed count and keeps the added label',
        (tester) async {
      final (apiClient, _) = await _fakeApiClient();

      await tester.pumpWidget(_wrap(
        apiClient,
        const InspectionFormScreen(apiaryId: 1, hive: _hive),
      ));
      await tester.pumpAndSettle();

      await _tapAdd(tester, _drawnField, times: 3);
      expect(find.text('Added empty frames'), findsOneWidget);
      expect(find.text('+3'), findsOneWidget);
    });

    testWidgets('negative drawn taps are sent as frames_added_drawn',
        (tester) async {
      final (apiClient, adapter) = await _fakeApiClient();

      await tester.pumpWidget(_wrap(
        apiClient,
        const InspectionFormScreen(apiaryId: 1, hive: _hive),
      ));
      await tester.pumpAndSettle();

      await _tapRemove(tester, _drawnField, times: 5);

      await tester.tap(find.byIcon(Icons.check));
      await tester.pumpAndSettle();

      final request = adapter.requests.firstWhere(
        (r) => r.path.contains('/inspections') && r.method == 'POST',
      );
      expect(request.data['frames_added_drawn'], -5);
    });

    testWidgets('signed frame fields are each sent in the inspection payload',
        (tester) async {
      final (apiClient, adapter) = await _fakeApiClient();

      await tester.pumpWidget(_wrap(
        apiClient,
        const InspectionFormScreen(apiaryId: 1, hive: _hive),
      ));
      await tester.pumpAndSettle();

      await _tapAdd(tester, _drawnField, times: 5);
      await _tapRemove(tester, _foundationField, times: 2);

      await tester.tap(find.byIcon(Icons.check));
      await tester.pumpAndSettle();

      final request = adapter.requests.firstWhere(
        (r) => r.path.contains('/inspections') && r.method == 'POST',
      );
      expect(request.data['frames_added_drawn'], 5);
      expect(request.data['frames_added_foundation'], -2);
    });
  });

  group('InspectionFormScreen size validation', () {
    testWidgets('truncates notes input at 5000 characters', (tester) async {
      final (apiClient, _) = await _fakeApiClient();

      await tester.pumpWidget(_wrap(
        apiClient,
        const InspectionFormScreen(apiaryId: 1, hive: _hive),
      ));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, 'Notes'),
        'a' * 5010,
      );
      await tester.pump();

      final field = tester.widget<TextFormField>(
        find.widgetWithText(TextFormField, 'Notes'),
      );
      expect(field.controller!.text.length, 5000);
    });

    testWidgets('truncates queen cells count input at 2 characters',
        (tester) async {
      final (apiClient, _) = await _fakeApiClient();

      await tester.pumpWidget(_wrap(
        apiClient,
        const InspectionFormScreen(apiaryId: 1, hive: _hive),
      ));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, 'Queen cells'),
        '1' * 10,
      );
      await tester.pump();

      final field = tester.widget<TextFormField>(
        find.widgetWithText(TextFormField, 'Queen cells'),
      );
      expect(field.controller!.text.length, 2);
    });
  });

  group('InspectionFormScreen upload progress', () {
    testWidgets(
      'shows a progress overlay instead of the remove button on a pending '
      'thumbnail while saving, and hides it once the upload settles',
      (tester) async {
        final (apiClient, _) = await _fakeApiClient();
        final imagePicker = _FakeImagePickerPlatform();
        ImagePickerPlatform.instance = imagePicker;

        await tester.pumpWidget(_wrap(
          apiClient,
          const InspectionFormScreen(apiaryId: 1, hive: _hive),
        ));
        await tester.pumpAndSettle();

        final addPhotoFinder = find.widgetWithIcon(
          IconButton,
          Icons.add_photo_alternate_outlined,
        );
        await tester.ensureVisible(addPhotoFinder);
        await tester.tap(addPhotoFinder);
        await tester.pumpAndSettle();
        await tester.tap(find.text(l10n.inspectionPhotoSourceGallery));
        await tester.pumpAndSettle();

        // Before submission the pending thumbnail shows a remove button and
        // no progress indicator.
        expect(find.byIcon(Icons.close), findsOneWidget);
        expect(find.byType(CircularProgressIndicator), findsNothing);

        await tester.tap(find.byIcon(Icons.check));
        // A single pump (not pumpAndSettle) catches the mid-upload frame
        // before the fake adapter's response resolves and the screen pops.
        await tester.pump();

        expect(find.byType(CircularProgressIndicator), findsWidgets);
        expect(find.byIcon(Icons.close), findsNothing);

        final addPhotoButton = tester.widget<IconButton>(addPhotoFinder);
        expect(addPhotoButton.onPressed, isNull);

        await tester.pumpAndSettle();
      },
    );

    testWidgets(
      'hides the existing-image remove button while saving',
      (tester) async {
        final (apiClient, adapter) = await _fakeApiClient();
        adapter.images = [
          {
            'id': 3,
            'inspection_id': 55,
            'mime_type': 'image/jpeg',
            'created_at': DateTime(2026, 1, 1).toIso8601String(),
          },
        ];
        final inspection = Inspection(
          id: 55,
          hiveId: 10,
          inspectedAt: DateTime(2025, 6, 1),
          queenSeen: 'seen',
          broodPattern: '',
          aggressiveness: '',
          queenAdded: false,
          notes: '',
        );

        await tester.pumpWidget(_wrap(
          apiClient,
          InspectionFormScreen(
            apiaryId: 1,
            hive: _hive,
            inspection: inspection,
          ),
        ));
        await tester.pumpAndSettle();

        expect(find.byIcon(Icons.close), findsOneWidget);

        await tester.tap(find.byIcon(Icons.check));
        await tester.pump();

        expect(find.byIcon(Icons.close), findsNothing);

        await tester.pumpAndSettle();
      },
    );
  });

  group('InspectionFormScreen queen status toggles', () {
    testWidgets(
        'marking queen added clears an already-selected queenless toggle',
        (tester) async {
      final (apiClient, _) = await _fakeApiClient();

      await tester.pumpWidget(_wrap(
        apiClient,
        const InspectionFormScreen(apiaryId: 1, hive: _hive),
      ));
      await tester.pumpAndSettle();

      final queenlessSwitch = find.descendant(
        of: find
            .ancestor(of: find.text(l10n.hiveQueenless), matching: find.byType(Row))
            .first,
        matching: find.byType(Switch),
      );
      final queenAddedSwitch = find.descendant(
        of: find
            .ancestor(
                of: find.text(l10n.inspectionQueenAdded), matching: find.byType(Row))
            .first,
        matching: find.byType(Switch),
      );

      await tester.ensureVisible(queenlessSwitch);
      await tester.tap(queenlessSwitch);
      await tester.pump();
      expect(tester.widget<Switch>(queenlessSwitch).value, isTrue);

      await tester.ensureVisible(queenAddedSwitch);
      await tester.tap(queenAddedSwitch);
      await tester.pump();

      expect(tester.widget<Switch>(queenAddedSwitch).value, isTrue);
      expect(tester.widget<Switch>(queenlessSwitch).value, isFalse);
      expect(tester.widget<Switch>(queenlessSwitch).onChanged, isNull);
    });
  });
}
