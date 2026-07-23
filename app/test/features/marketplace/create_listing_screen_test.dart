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
import 'package:app/features/marketplace/data/listing_model.dart';
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
  List<Map<String, dynamic>> honeyBatches = [];
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
    if (options.path.endsWith('/honey-batches') && options.method == 'GET') {
      return _json({'items': honeyBatches, 'total': honeyBatches.length});
    }
    if (options.path.endsWith('/listings') && options.method == 'POST') {
      if (failCreate) {
        throw DioException(requestOptions: options, message: 'create failed');
      }
      final data = options.data as Map<String, dynamic>;
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
        'honey_batch_id': data['honey_batch_id'],
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
    if (options.path.contains('/images') && options.method == 'DELETE') {
      return _json({});
    }
    if (RegExp(r'/listings/\d+$').hasMatch(options.path) &&
        options.method == 'PATCH') {
      final data = options.data as Map<String, dynamic>;
      return _json({
        'id': 99,
        'user_id': 1,
        'title': data['title'],
        'description': data['description'],
        'category': data['category'],
        'price': data['price'],
        'quantity': data['quantity'],
        'address': data['address'],
        'contact_phone': data['contact_phone'],
        'contact_email': data['contact_email'],
        'is_hidden': false,
        'created_at': DateTime(2026, 1, 1).toIso8601String(),
        'updated_at': DateTime(2026, 1, 1).toIso8601String(),
        'images': [],
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
  final apiClient = ApiClient(
    storage: TokenStorage(prefs),
    baseUrl: 'http://test',
  );
  apiClient.dio.httpClientAdapter = adapter;
  return apiClient;
}

Widget _wrap(ApiClient apiClient, Widget child) =>
    RepositoryProvider<ApiClient>.value(
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
  Listing? existingListing,
}) => RepositoryProvider<ApiClient>.value(
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
                MaterialPageRoute(
                  builder: (_) =>
                      CreateListingScreen(existingListing: existingListing),
                ),
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

Listing _existingListing({int id = 5, List<ListingImage> images = const []}) =>
    Listing(
      id: id,
      userId: 1,
      title: 'Old Title',
      description: 'Old description.',
      category: 'HONEY',
      price: 12.5,
      quantity: '3 jars',
      address: 'Krakow',
      lat: 50.0647,
      lng: 19.945,
      contactPhone: '123456789',
      contactEmail: 'seller@example.com',
      isHidden: false,
      createdAt: DateTime(2026, 1, 1),
      updatedAt: DateTime(2026, 1, 1),
      images: images,
    );

void main() {
  late AppLocalizations l10n;
  setUpAll(() async {
    l10n = await AppLocalizations.delegate.load(const Locale('en'));
  });

  group('CreateListingScreen validation', () {
    testWidgets('shows title required error when title is left empty', (
      tester,
    ) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      tester.state<FormState>(find.byType(Form)).validate();
      await tester.pump();

      expect(find.text(l10n.marketplaceFieldTitleRequired), findsOneWidget);
    });

    testWidgets('shows category required error when category is not selected', (
      tester,
    ) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldTitle),
        'Wildflower Honey',
      );
      final saveFinder = find.byIcon(Icons.check);
      await tester.ensureVisible(saveFinder);
      await tester.tap(saveFinder);
      await tester.pumpAndSettle();

      expect(find.text(l10n.marketplaceFieldCategoryRequired), findsOneWidget);
      expect(find.byType(SnackBar), findsNothing);
    });

    testWidgets('truncates description input at 500 characters', (
      tester,
    ) async {
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

    testWidgets('truncates title input at 150 characters', (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldTitle),
        'a' * 160,
      );
      await tester.pump();

      final field = tester.widget<TextFormField>(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldTitle),
      );
      expect(field.controller!.text.length, 150);
    });

    testWidgets('truncates quantity input at 50 characters', (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceQuantityLabel),
        'a' * 60,
      );
      await tester.pump();

      final field = tester.widget<TextFormField>(
        find.widgetWithText(TextFormField, l10n.marketplaceQuantityLabel),
      );
      expect(field.controller!.text.length, 50);
    });

    testWidgets('truncates address input at 150 characters', (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldAddress),
        'a' * 160,
      );
      await tester.pump();

      final field = tester.widget<TextFormField>(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldAddress),
      );
      expect(field.controller!.text.length, 150);
    });

    testWidgets('truncates phone input at 20 characters', (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldPhone),
        '1' * 30,
      );
      await tester.pump();

      final field = tester.widget<TextFormField>(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldPhone),
      );
      expect(field.controller!.text.length, 20);
    });

    testWidgets('truncates email input at 150 characters', (tester) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldEmail),
        '${'a' * 150}@example.com',
      );
      await tester.pump();

      final field = tester.widget<TextFormField>(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldEmail),
      );
      expect(field.controller!.text.length, 150);
    });

    testWidgets('shows price invalid error for non-numeric price', (
      tester,
    ) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      // '.' alone passes the price input formatter (digits, optional dot,
      // up to 2 decimals) but is not a parseable double.
      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldPrice),
        '.',
      );
      tester.state<FormState>(find.byType(Form)).validate();
      await tester.pump();

      expect(find.text(l10n.marketplaceFieldPriceInvalid), findsOneWidget);
    });

    testWidgets(
      'shows price too large error for a price at or above 100,000,000',
      (tester) async {
        final adapter = _RecordingHttpClientAdapter();
        final apiClient = await _fakeApiClient(adapter);

        await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
        await tester.pumpAndSettle();

        await tester.enterText(
          find.widgetWithText(TextFormField, l10n.marketplaceFieldPrice),
          '883154044',
        );
        tester.state<FormState>(find.byType(Form)).validate();
        await tester.pump();

        expect(find.text(l10n.marketplaceFieldPriceTooLarge), findsOneWidget);
      },
    );

    testWidgets('price input formatter rejects non-numeric characters', (
      tester,
    ) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldPrice),
        'not-a-number',
      );
      await tester.pump();

      final field = tester.widget<TextFormField>(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldPrice),
      );
      expect(field.controller!.text, isEmpty);
    });

    testWidgets('price input formatter accepts a valid two-decimal price', (
      tester,
    ) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldPrice),
        '12.50',
      );
      await tester.pump();

      final field = tester.widget<TextFormField>(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldPrice),
      );
      expect(field.controller!.text, '12.50');
    });

    testWidgets('price input formatter rejects a third decimal digit', (
      tester,
    ) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      final priceFinder = find.widgetWithText(
        TextFormField,
        l10n.marketplaceFieldPrice,
      );

      await tester.enterText(priceFinder, '12.99');
      await tester.pump();
      // Appending a third decimal digit to the already-valid '12.99' is
      // rejected, leaving the field unchanged.
      await tester.enterText(priceFinder, '12.999');
      await tester.pump();

      final field = tester.widget<TextFormField>(priceFinder);
      expect(field.controller!.text, '12.99');
    });

    testWidgets('shows price required error when price is left empty', (
      tester,
    ) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      tester.state<FormState>(find.byType(Form)).validate();
      await tester.pump();

      expect(find.text(l10n.marketplaceFieldPriceRequired), findsOneWidget);
      expect(find.text(l10n.marketplaceFieldPriceInvalid), findsNothing);
    });

    testWidgets('shows phone invalid error for malformed phone number', (
      tester,
    ) async {
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

    testWidgets('does not flag phone invalid when phone is left empty', (
      tester,
    ) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      tester.state<FormState>(find.byType(Form)).validate();
      await tester.pump();

      expect(find.text(l10n.marketplaceFieldPhoneInvalid), findsNothing);
    });

    testWidgets('shows invalid email error for malformed email', (
      tester,
    ) async {
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
      'fields above it',
      (tester) async {
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
        // The email field itself autovalidates on its own interaction, so its
        // own error is expected to show; only the OTHER fields must stay
        // untouched.
        expect(find.text(l10n.authInvalidEmail), findsOneWidget);
      },
    );

    testWidgets(
      'clears the title error live once a valid title is typed, without '
      'pressing save again',
      (tester) async {
        final adapter = _RecordingHttpClientAdapter();
        final apiClient = await _fakeApiClient(adapter);

        await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
        await tester.pumpAndSettle();

        final saveFinder = find.byIcon(Icons.check);
        await tester.ensureVisible(saveFinder);
        await tester.tap(saveFinder);
        await tester.pumpAndSettle();

        expect(find.text(l10n.marketplaceFieldTitleRequired), findsOneWidget);

        await tester.enterText(
          find.widgetWithText(TextFormField, l10n.marketplaceFieldTitle),
          'Wildflower Honey',
        );
        await tester.pump();

        expect(find.text(l10n.marketplaceFieldTitleRequired), findsNothing);
      },
    );
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

    testWidgets('is shown with the apiary options when apiaries exist', (
      tester,
    ) async {
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

  group('CreateListingScreen honey batch attach', () {
    Map<String, dynamic> confirmedBatch({int id = 1}) => {
      'id': id,
      'verification_token': 'tok$id',
      'verification_url': 'https://example.com/verify/tok$id',
      'gathering_date': DateTime(2026, 1, 1).toIso8601String(),
      'amount_grams': 2500,
      'processing_method': 'cold_extracted',
      'honey_type': 'Wildflower',
      'pdf_filename': 'lab.pdf',
      'pdf_file_hash': 'hash',
      'created_at': DateTime(2026, 1, 1).toIso8601String(),
      'updated_at': DateTime(2026, 1, 1).toIso8601String(),
      'certification': {'status': 'confirmed'},
    };

    testWidgets('is hidden until the HONEY category is selected', (
      tester,
    ) async {
      final adapter = _RecordingHttpClientAdapter()
        ..honeyBatches = [confirmedBatch()];
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      expect(find.text(l10n.marketplaceHoneyBatchAttachLabel), findsNothing);

      await tester.tap(find.byType(DropdownButtonFormField<String>));
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.marketplaceCategoryHoney).last);
      await tester.pumpAndSettle();

      expect(
        find.text(l10n.marketplaceHoneyBatchAttachLabel),
        findsWidgets,
      );
    });

    testWidgets('shows "none available" when the user has no confirmed batches', (
      tester,
    ) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      await tester.tap(find.byType(DropdownButtonFormField<String>));
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.marketplaceCategoryHoney).last);
      await tester.pumpAndSettle();

      expect(find.text(l10n.marketplaceHoneyBatchNoneAvailable), findsOneWidget);
    });

    testWidgets('threads the selected honey batch id into the create request', (
      tester,
    ) async {
      final adapter = _RecordingHttpClientAdapter()
        ..honeyBatches = [confirmedBatch(id: 7)];
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

      await tester.tap(find.byType(DropdownButtonFormField<int?>));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Wildflower · 2.5 kg').last);
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldPrice),
        '42.50',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.marketplaceFieldPhone),
        '123456789',
      );
      _setLocationFields(tester, l10n);
      await tester.pump();

      final saveFinder = find.byIcon(Icons.check);
      await tester.ensureVisible(saveFinder);
      await tester.tap(saveFinder);
      await tester.pumpAndSettle();

      final createRequests = adapter.requests.where(
        (r) => r.path.endsWith('/listings') && r.method == 'POST',
      );
      expect(createRequests, hasLength(1));
      expect(
        (createRequests.first.data as Map<String, dynamic>)['honey_batch_id'],
        7,
      );
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
      _setLocationFields(tester, l10n);
      await tester.pump();
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

        final addPhotoFinder = find.widgetWithIcon(
          IconButton,
          Icons.add_photo_alternate_outlined,
        );
        await tester.ensureVisible(addPhotoFinder);
        await tester.tap(addPhotoFinder);
        await tester.pumpAndSettle();
        await tester.tap(find.text(l10n.marketplacePhotoSourceGallery));
        await tester.pumpAndSettle();

        final saveFinder = find.byIcon(Icons.check);
        await tester.ensureVisible(saveFinder);
        await tester.tap(saveFinder);
        await tester.pumpAndSettle();

        expect(result, isTrue);
        expect(find.byType(CreateListingScreen), findsNothing);

        final createRequests = adapter.requests.where(
          (r) => r.path.endsWith('/listings') && r.method == 'POST',
        );
        final uploadRequests = adapter.requests.where(
          (r) => r.path.contains('/images') && r.method == 'POST',
        );
        expect(createRequests, hasLength(1));
        expect(uploadRequests, hasLength(1));
      },
    );

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

        final saveFinder = find.byIcon(Icons.check);
        await tester.ensureVisible(saveFinder);
        await tester.tap(saveFinder);
        await tester.pumpAndSettle();

        expect(find.text(l10n.generalError), findsOneWidget);
        expect(find.byType(SnackBar), findsNothing);
        expect(find.byType(CreateListingScreen), findsOneWidget);
        expect(result, isNull);

        final saveButton = tester.widget<IconButton>(
          find.widgetWithIcon(IconButton, Icons.check),
        );
        expect(saveButton.onPressed, isNotNull);
      },
    );

    testWidgets(
      'blocks submission and shows an error when neither phone nor email '
      'is provided',
      (tester) async {
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

        final saveFinder = find.byIcon(Icons.check);
        await tester.ensureVisible(saveFinder);
        await tester.tap(saveFinder);
        await tester.pumpAndSettle();

        expect(find.text(l10n.marketplaceContactRequired), findsOneWidget);
        expect(find.byType(SnackBar), findsNothing);
        final createRequests = adapter.requests.where(
          (r) => r.path.endsWith('/listings') && r.method == 'POST',
        );
        expect(createRequests, isEmpty);

        await tester.enterText(
          find.widgetWithText(TextFormField, l10n.marketplaceFieldPhone),
          '+15551234567',
        );
        await tester.pumpAndSettle();

        expect(find.text(l10n.marketplaceContactRequired), findsNothing);
      },
    );

    testWidgets(
      'blocks submission and shows an error when no location has been picked',
      (tester) async {
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
        await tester.enterText(
          find.widgetWithText(TextFormField, l10n.marketplaceFieldPhone),
          '123456789',
        );

        final saveFinder = find.byIcon(Icons.check);
        await tester.ensureVisible(saveFinder);
        await tester.tap(saveFinder);
        await tester.pumpAndSettle();

        expect(find.text(l10n.marketplaceLocationRequired), findsOneWidget);
        final createRequests = adapter.requests.where(
          (r) => r.path.endsWith('/listings') && r.method == 'POST',
        );
        expect(createRequests, isEmpty);

        _setLocationFields(tester, l10n);
        await tester.pump();
        await tester.tap(saveFinder);
        await tester.pumpAndSettle();

        expect(find.text(l10n.marketplaceLocationRequired), findsNothing);
        final createRequestsAfter = adapter.requests.where(
          (r) => r.path.endsWith('/listings') && r.method == 'POST',
        );
        expect(createRequestsAfter, hasLength(1));
      },
    );
  });

  group('CreateListingScreen editing', () {
    testWidgets('prefills fields and shows the edit title when editing', (
      tester,
    ) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);
      final listing = _existingListing();

      await tester.pumpWidget(
        _wrap(apiClient, CreateListingScreen(existingListing: listing)),
      );
      await tester.pumpAndSettle();

      expect(find.text(l10n.marketplaceEditScreenTitle), findsOneWidget);
      expect(find.text('Old Title'), findsOneWidget);
      expect(find.text('Old description.'), findsOneWidget);
      expect(find.text('12.50'), findsOneWidget);
      expect(find.text('3 jars'), findsOneWidget);
    });

    testWidgets('submits via PATCH and pops true without creating a listing', (
      tester,
    ) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);
      final listing = _existingListing();
      bool? result;

      await tester.pumpWidget(
        _wrapWithNavigator(
          apiClient,
          existingListing: listing,
          onResult: (r) => result = r,
        ),
      );
      await tester.tap(find.text('open'));
      await tester.pumpAndSettle();

      final saveFinder = find.byIcon(Icons.check);
      await tester.ensureVisible(saveFinder);
      await tester.tap(saveFinder);
      await tester.pumpAndSettle();

      expect(result, isTrue);
      final patchRequests = adapter.requests.where(
        (r) =>
            r.path.endsWith('/listings/${listing.id}') && r.method == 'PATCH',
      );
      final createRequests = adapter.requests.where(
        (r) => r.path.endsWith('/listings') && r.method == 'POST',
      );
      expect(patchRequests, hasLength(1));
      expect(createRequests, isEmpty);
    });

    testWidgets(
      'shows existing images and deletes one via the repository on tap',
      (tester) async {
        final adapter = _RecordingHttpClientAdapter();
        final apiClient = await _fakeApiClient(adapter);
        final image = ListingImage(
          id: 3,
          listingId: 5,
          url: '/uploads/a.jpg',
          displayOrder: 0,
          createdAt: DateTime(2026, 1, 1),
        );
        final listing = _existingListing(images: [image]);

        await tester.pumpWidget(
          _wrap(apiClient, CreateListingScreen(existingListing: listing)),
        );
        await tester.pumpAndSettle();

        expect(
          find.text('${l10n.marketplacePhotosLabel}  1/3'),
          findsOneWidget,
        );
        expect(find.byIcon(Icons.close), findsOneWidget);

        await tester.ensureVisible(find.byIcon(Icons.close));
        await tester.tap(find.byIcon(Icons.close));
        await tester.pumpAndSettle();

        expect(
          find.textContaining(l10n.marketplacePhotosLabel),
          findsNothing,
        );
        final deleteRequests = adapter.requests.where(
          (r) =>
              r.path.endsWith('/listings/${listing.id}/images/${image.id}') &&
              r.method == 'DELETE',
        );
        expect(deleteRequests, hasLength(1));
      },
    );
  });

  group('CreateListingScreen photo picker', () {
    testWidgets('caps picked photos at three and disables adding more', (
      tester,
    ) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);
      final imagePicker = _FakeImagePickerPlatform();
      ImagePickerPlatform.instance = imagePicker;

      await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
      await tester.pumpAndSettle();

      final addPhotoFinder = find.widgetWithIcon(
        IconButton,
        Icons.add_photo_alternate_outlined,
      );
      for (var i = 0; i < 3; i++) {
        await tester.ensureVisible(addPhotoFinder);
        await tester.tap(addPhotoFinder);
        await tester.pumpAndSettle();
        await tester.tap(find.text(l10n.marketplacePhotoSourceGallery));
        await tester.pumpAndSettle();
      }

      expect(find.text('${l10n.marketplacePhotosLabel}  3/3'), findsOneWidget);
      expect(imagePicker.pickCount, 3);

      final addPhotoButton = tester.widget<IconButton>(
        find.widgetWithIcon(IconButton, Icons.add_photo_alternate_outlined),
      );
      expect(addPhotoButton.onPressed, isNull);

      await tester.ensureVisible(addPhotoFinder);
      await tester.tap(addPhotoFinder);
      await tester.pumpAndSettle();

      expect(imagePicker.pickCount, 3);
    });
  });

  group('CreateListingScreen upload progress', () {
    testWidgets(
      'shows a progress overlay instead of the remove button on a pending '
      'thumbnail while saving, and hides it once the upload settles',
      (tester) async {
        final adapter = _RecordingHttpClientAdapter();
        final apiClient = await _fakeApiClient(adapter);
        final imagePicker = _FakeImagePickerPlatform();
        ImagePickerPlatform.instance = imagePicker;

        await tester.pumpWidget(_wrap(apiClient, const CreateListingScreen()));
        await tester.pumpAndSettle();

        await _fillRequiredFieldsStandalone(tester, l10n);

        final addPhotoFinder = find.widgetWithIcon(
          IconButton,
          Icons.add_photo_alternate_outlined,
        );
        await tester.ensureVisible(addPhotoFinder);
        await tester.tap(addPhotoFinder);
        await tester.pumpAndSettle();
        await tester.tap(find.text(l10n.marketplacePhotoSourceGallery));
        await tester.pumpAndSettle();

        // Before submission the pending thumbnail shows a remove button and
        // no progress indicator.
        expect(find.byIcon(Icons.close), findsOneWidget);
        expect(find.byType(CircularProgressIndicator), findsNothing);

        final saveFinder = find.byIcon(Icons.check);
        await tester.ensureVisible(saveFinder);
        await tester.tap(saveFinder);
        // A single pump (not pumpAndSettle) catches the mid-upload frame
        // before the fake adapter's response resolves and the screen pops.
        await tester.pump();

        expect(find.byType(CircularProgressIndicator), findsWidgets);
        expect(find.byIcon(Icons.close), findsNothing);

        final addPhotoButton = tester.widget<IconButton>(addPhotoFinder);
        expect(addPhotoButton.onPressed, isNull);

        await tester.pumpAndSettle();
        expect(find.byType(CreateListingScreen), findsNothing);
      },
    );

    testWidgets('hides the existing-image remove button while saving', (
      tester,
    ) async {
      final adapter = _RecordingHttpClientAdapter();
      final apiClient = await _fakeApiClient(adapter);
      final image = ListingImage(
        id: 3,
        listingId: 5,
        url: '/uploads/a.jpg',
        displayOrder: 0,
        createdAt: DateTime(2026, 1, 1),
      );
      final listing = _existingListing(images: [image]);

      await tester.pumpWidget(
        _wrap(apiClient, CreateListingScreen(existingListing: listing)),
      );
      await tester.pumpAndSettle();

      expect(find.byIcon(Icons.close), findsOneWidget);

      final saveFinder = find.byIcon(Icons.check);
      await tester.ensureVisible(saveFinder);
      await tester.tap(saveFinder);
      await tester.pump();

      expect(find.byIcon(Icons.close), findsNothing);

      await tester.pumpAndSettle();
    });
  });
}

Future<void> _fillRequiredFieldsStandalone(
  WidgetTester tester,
  AppLocalizations l10n,
) async {
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
  _setLocationFields(tester, l10n);
  await tester.pump();
}

/// Sets the (disabled, picker-only) latitude/longitude fields directly via
/// their controllers — mirrors what `_setLocation` does internally when the
/// user taps GPS or picks a point on the map, which aren't drivable here.
void _setLocationFields(WidgetTester tester, AppLocalizations l10n) {
  tester
          .widget<TextFormField>(
            find.widgetWithText(TextFormField, l10n.marketplaceFieldLatitude),
          )
          .controller!
          .text =
      '52.229700';
  tester
          .widget<TextFormField>(
            find.widgetWithText(TextFormField, l10n.marketplaceFieldLongitude),
          )
          .controller!
          .text =
      '21.012200';
}
