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
import 'package:app/features/marketplace/view/create_listing_screen.dart';
import 'package:app/l10n/app_localizations.dart';

/// A valid 1x1 transparent PNG, so `Image.memory` can decode the picked
/// files' bytes in `_PendingThumbnail` without throwing.
final Uint8List _kOnePixelPng = base64Decode(
  'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mNk+A8AAQUBAScY'
  '42YAAAAASUVORK5CYII=',
);

/// A fake image_picker platform that returns a new in-memory [XFile] on every
/// pick, so photo-picker flows can be exercised without a real device/plugin.
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

/// Records every request and serves canned JSON for the endpoints
/// CreateListingScreen and its dependencies (ProfileIconButton,
/// ApiaryRepository, ListingRepository) call.
class _RecordingHttpClientAdapter implements HttpClientAdapter {
  final List<RequestOptions> requests = [];
  List<Map<String, dynamic>> apiaries = [];
  bool failCreate = false;
  int _nextImageId = 1;

  @override
  void close({bool force = false}) {}

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    requests.add(options);

    if (options.path.contains('/invitations/count')) {
      return _json({'count': 0});
    }
    if (options.path.endsWith('/apiaries') && options.method == 'GET') {
      return _json(apiaries);
    }
    if (options.path.endsWith('/listings') && options.method == 'POST') {
      if (failCreate) {
        throw DioException(requestOptions: options, message: 'create failed');
      }
      return _json({
        'id': 42,
        'user_id': 1,
        'title': 'Wildflower Honey',
        'description': '',
        'category': 'HONEY',
        'price': null,
        'quantity': '',
        'address': '',
        'contact_phone': '',
        'contact_email': '',
        'is_hidden': false,
        'created_at': DateTime(2026, 1, 1).toIso8601String(),
        'updated_at': DateTime(2026, 1, 1).toIso8601String(),
        'images': [],
      });
    }
    if (options.path.contains('/images') && options.method == 'POST') {
      final id = _nextImageId++;
      return _json({
        'id': id,
        'listing_id': 42,
        'url': 'http://test/img$id.jpg',
        'display_order': id,
        'created_at': DateTime(2026, 1, 1).toIso8601String(),
      });
    }
    return _json({});
  }

  ResponseBody _json(Object? data) => ResponseBody.fromString(
        jsonEncode(data),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
}

Future<ApiClient> _fakeApiClient(_RecordingHttpClientAdapter adapter) async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  final apiClient = ApiClient(storage: TokenStorage(prefs), baseUrl: 'http://test');
  apiClient.dio.httpClientAdapter = adapter;
  return apiClient;
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

/// Wraps CreateListingScreen behind a launcher button so pushed/popped
/// results can be observed like a real caller would see them.
Widget _wrapWithNavigator(
  ApiClient apiClient, {
  required ValueChanged<bool?> onResult,
}) =>
    RepositoryProvider<ApiClient>.value(
      value: apiClient,
      child: MaterialApp(
        localizationsDelegates: AppLocalizations.localizationsDelegates,
        supportedLocales: AppLocalizations.supportedLocales,
        locale: const Locale('en'),
        home: Builder(
          builder: (context) => Scaffold(
            body: Center(
              child: ElevatedButton(
                onPressed: () async {
                  final result = await Navigator.of(context).push<bool>(
                    MaterialPageRoute(builder: (_) => const CreateListingScreen()),
                  );
                  onResult(result);
                },
                child: const Text('open'),
              ),
            ),
          ),
        ),
      ),
    );

void main() {
  late AppLocalizations l10n;
  setUpAll(() async {
    l10n = await AppLocalizations.delegate.load(const Locale('en'));
  });

  group('CreateListingScreen validation', () {
    testWidgets('shows title required error when title is left empty',
        (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      tester.state<FormState>(find.byType(Form)).validate();
      await tester.pump();

      expect(find.text(l10n.marketplaceFieldTitleRequired), findsOneWidget);
    });

    testWidgets('shows category required error when category is not selected',
        (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldTitle),
        'Wildflower Honey',
      );
      final saveFinder = find.widgetWithText(ElevatedButton, l10n.generalSave);
      await tester.ensureVisible(saveFinder);
      await tester.tap(saveFinder);
      await tester.pumpAndSettle();

      expect(find.text(l10n.marketplaceFieldCategoryRequired), findsOneWidget);
      expect(find.byType(SnackBar), findsNothing);
    });

    testWidgets('truncates description input at 500 characters',
        (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceDescriptionLabel),
        'a' * 510,
      );
      await tester.pump();

      final field = tester.widget<TextFormField>(
        find.widgetWithText(TextFormField, l10n.marketplaceDescriptionLabel),
      );
      expect(field.controller!.text.length, 500);
      expect(find.text('500/500'), findsOneWidget);
    });

    testWidgets('shows price invalid error for non-numeric price',
        (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldPrice),
        'not-a-number',
      );
      tester.state<FormState>(find.byType(Form)).validate();
      await tester.pump();

      expect(find.text(l10n.marketplaceFieldPriceInvalid), findsOneWidget);
    });

    testWidgets('shows price required error when price is left empty',
        (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      tester.state<FormState>(find.byType(Form)).validate();
      await tester.pump();

      expect(find.text(l10n.marketplaceFieldPriceRequired), findsOneWidget);
      expect(find.text(l10n.marketplaceFieldPriceInvalid), findsNothing);
    });

    testWidgets('shows phone invalid error for malformed phone number',
        (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldPhone),
        '12345',
      );
      tester.state<FormState>(find.byType(Form)).validate();
      await tester.pump();

      expect(find.text(l10n.marketplaceFieldPhoneInvalid), findsOneWidget);
    });

    testWidgets('does not flag phone invalid when phone is left empty',
        (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      tester.state<FormState>(find.byType(Form)).validate();
      await tester.pump();

      expect(find.text(l10n.marketplaceFieldPhoneInvalid), findsNothing);
    });

    testWidgets('shows invalid email error for malformed email',
        (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldEmail),
        'not-an-email',
      );
      tester.state<FormState>(find.byType(Form)).validate();
      await tester.pump();

      expect(find.text(l10n.authInvalidEmail), findsOneWidget);
    });

    testWidgets(
        'typing garbage into the email field alone does not flag untouched '
        'fields above it', (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldEmail),
        'not-an-email',
      );
      await tester.pump();

      expect(find.text(l10n.marketplaceFieldTitleRequired), findsNothing);
      expect(find.text(l10n.marketplaceFieldCategoryRequired), findsNothing);
      expect(find.text(l10n.marketplaceFieldPriceRequired), findsNothing);
      expect(find.text(l10n.authInvalidEmail), findsNothing);
    });
  });

  group('CreateListingScreen apiary dropdown', () {
    testWidgets('is hidden when the user has no apiaries', (tester) async {
      final adapter = _RecordingHttpClientAdapter()..apiaries = [];
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      expect(find.text(l10n.marketplaceApiaryLabel), findsNothing);
      expect(find.byType(DropdownButtonFormField<int?>), findsNothing);
    });

    testWidgets('is shown with the apiary options when apiaries exist',
        (tester) async {
      final adapter = _RecordingHttpClientAdapter()
        ..apiaries = [
          {
            'id': 1,
            'name': 'Home apiary',
            'lat': null,
            'lng': null,
            'grid_rows': 2,
            'grid_cols': 2,
            'hive_count': 0,
            'user_role': 'OWNER',
          },
        ];
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      expect(find.text(l10n.marketplaceApiaryLabel), findsOneWidget);
      expect(find.byType(DropdownButtonFormField<int?>), findsOneWidget);
    });
  });

  group('CreateListingScreen submit', () {
    Future<void> _fillRequiredFields(WidgetTester tester) async {
      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldTitle),
        'Wildflower Honey',
      );
      await tester.tap(find.byType(DropdownButtonFormField<String>));
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.marketplaceCategoryHoney).last);
      await tester.pumpAndSettle();
      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldPrice),
        '42.50',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldPhone),
        '123456789',
      );
    }

    testWidgets(
        'creates the listing, uploads each picked photo, then pops true',
        (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);
      final imagePicker = _FakeImagePickerPlatform();
      ImagePickerPlatform.instance = imagePicker;
      bool? result;

      await tester.pumpWidget(
        _wrapWithNavigator(apiClient, onResult: (r) => result = r),
      );
      await tester.tap(find.text('open'));
      await tester.pumpAndSettle();

      await _fillRequiredFields(tester);

      final addPhotoFinder =
          find.widgetWithText(TextButton, l10n.marketplaceAddPhoto);
      await tester.ensureVisible(addPhotoFinder);
      await tester.tap(addPhotoFinder);
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.marketplacePhotoSourceGallery));
      await tester.pumpAndSettle();

      final saveFinder = find.widgetWithText(ElevatedButton, l10n.generalSave);
      await tester.ensureVisible(saveFinder);
      await tester.tap(saveFinder);
      await tester.pumpAndSettle();

      expect(result, isTrue);
      expect(find.byType(CreateListingScreen), findsNothing);

      final createRequests = adapter.requests
          .where((r) => r.path.endsWith('/listings') && r.method == 'POST');
      final uploadRequests = adapter.requests
          .where((r) => r.path.contains('/images') && r.method == 'POST');
      expect(createRequests, hasLength(1));
      expect(uploadRequests, hasLength(1));
    });

    testWidgets(
        'shows an inline error and re-enables the form when creation fails',
        (tester) async {
      final adapter = _RecordingHttpClientAdapter()..failCreate = true;
      final apiClient = await _fakeApiClient(adapter);
      bool? result;

      await tester.pumpWidget(
        _wrapWithNavigator(apiClient, onResult: (r) => result = r),
      );
      await tester.tap(find.text('open'));
      await tester.pumpAndSettle();

      await _fillRequiredFields(tester);

      final saveFinder = find.widgetWithText(ElevatedButton, l10n.generalSave);
      await tester.ensureVisible(saveFinder);
      await tester.tap(saveFinder);
      await tester.pumpAndSettle();

      expect(find.text(l10n.generalError), findsOneWidget);
      expect(find.byType(SnackBar), findsNothing);
      expect(find.byType(CreateListingScreen), findsOneWidget);
      expect(result, isNull);

      final saveButton = tester.widget<ElevatedButton>(
        find.widgetWithText(ElevatedButton, l10n.generalSave),
      );
      expect(saveButton.onPressed, isNotNull);
    });

    testWidgets(
        'blocks submission and shows an error when neither phone nor email '
        'is provided', (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldTitle),
        'Wildflower Honey',
      );
      await tester.tap(find.byType(DropdownButtonFormField<String>));
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.marketplaceCategoryHoney).last);
      await tester.pumpAndSettle();
      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldPrice),
        '42.50',
      );

      final saveFinder = find.widgetWithText(ElevatedButton, l10n.generalSave);
      await tester.ensureVisible(saveFinder);
      await tester.tap(saveFinder);
      await tester.pumpAndSettle();

      expect(find.text(l10n.marketplaceContactRequired), findsWidgets);
      expect(find.byType(SnackBar), findsNothing);
      final createRequests = adapter.requests
          .where((r) => r.path.endsWith('/listings') && r.method == 'POST');
      expect(createRequests, isEmpty);
    });
  });

  group('CreateListingScreen photo picker', () {
    testWidgets('caps picked photos at three and disables adding more',
        (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);
      final imagePicker = _FakeImagePickerPlatform();
      ImagePickerPlatform.instance = imagePicker;

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      final addPhotoFinder =
          find.widgetWithText(TextButton, l10n.marketplaceAddPhoto);
      for (var i = 0; i < 3; i++) {
        await tester.ensureVisible(addPhotoFinder);
        await tester.tap(addPhotoFinder);
        await tester.pumpAndSettle();
        await tester.tap(find.text(l10n.marketplacePhotoSourceGallery));
        await tester.pumpAndSettle();
      }

      expect(find.text('${l10n.marketplacePhotosLabel}  3/3'), findsOneWidget);
      expect(imagePicker.pickCount, 3);

      final addPhotoButton = tester.widget<TextButton>(
        find.widgetWithText(TextButton, l10n.marketplaceAddPhoto),
      );
      expect(addPhotoButton.onPressed, isNull);

      await tester.ensureVisible(addPhotoFinder);
      await tester.tap(addPhotoFinder);
      await tester.pumpAndSettle();

      expect(imagePicker.pickCount, 3);
    });
  });
}
