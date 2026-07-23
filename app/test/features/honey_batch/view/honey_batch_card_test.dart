import 'dart:convert';
import 'dart:typed_data';

import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'package:app/core/api/api_client.dart';
import 'package:app/core/storage/token_storage.dart';
import 'package:app/features/honey_batch/cubit/honey_batches_cubit.dart';
import 'package:app/features/honey_batch/data/honey_batch_certification_model.dart';
import 'package:app/features/honey_batch/data/honey_batch_certification_request_model.dart';
import 'package:app/features/honey_batch/data/honey_batch_model.dart';
import 'package:app/features/honey_batch/data/honey_batch_repository.dart';
import 'package:app/features/honey_batch/data/processing_method.dart';
import 'package:app/features/honey_batch/view/create_honey_batch_screen.dart';
import 'package:app/features/honey_batch/view/honey_batches_home_screen.dart';
import 'package:app/l10n/app_localizations.dart';

class MockHoneyBatchRepository extends Mock implements HoneyBatchRepository {}

/// Solves the confirmation math puzzle (see `showDeleteDialog` /
/// `withPuzzle: true`) by reading the displayed "a + b = " prompt, typing
/// the sum, and tapping the button labelled [confirmLabel].
Future<void> _solvePuzzle(
  WidgetTester tester,
  String confirmLabel,
) async {
  final prompt = tester
      .widgetList<Text>(find.byType(Text))
      .firstWhere((t) => RegExp(r'^\d+ \+ \d+ = $').hasMatch(t.data ?? ''));
  final match = RegExp(r'^(\d+) \+ (\d+) = $').firstMatch(prompt.data!)!;
  final sum = int.parse(match.group(1)!) + int.parse(match.group(2)!);
  await tester.enterText(find.byType(TextField), '$sum');
  await tester.tap(find.text(confirmLabel).last);
  await tester.pumpAndSettle();
}

/// Returns `[]` for any request except a PATCH to /honey-batches/:id, which
/// returns [patchResponse] — lets tests exercise a real end-to-end submit
/// (e.g. the edit form) through a widget tree that also contains
/// ProfileIconButton, which fires its own GET on init.
class _FakeApiAdapter implements HttpClientAdapter {
  final Map<String, dynamic>? patchResponse;
  RequestOptions? lastOptions;

  _FakeApiAdapter({this.patchResponse});

  @override
  Future<ResponseBody> fetch(
    RequestOptions options,
    Stream<Uint8List>? requestStream,
    Future<void>? cancelFuture,
  ) async {
    lastOptions = options;
    if (options.method == 'PATCH' && patchResponse != null) {
      return ResponseBody.fromString(
        jsonEncode(patchResponse),
        200,
        headers: {
          Headers.contentTypeHeader: [Headers.jsonContentType],
        },
      );
    }
    return ResponseBody.fromString(
      jsonEncode(<Map<String, dynamic>>[]),
      200,
      headers: {
        Headers.contentTypeHeader: [Headers.jsonContentType],
      },
    );
  }

  @override
  void close({bool force = false}) {}
}

Future<ApiClient> _fakeApiClient({Map<String, dynamic>? patchResponse}) async {
  SharedPreferences.setMockInitialValues({});
  final prefs = await SharedPreferences.getInstance();
  final apiClient =
      ApiClient(storage: TokenStorage(prefs), baseUrl: 'http://test');
  apiClient.dio.httpClientAdapter =
      _FakeApiAdapter(patchResponse: patchResponse);
  return apiClient;
}

HoneyBatchModel _batch({
  int id = 1,
  ProcessingMethod processingMethod = ProcessingMethod.raw,
  String honeyType = 'Wildflower',
  int amountGrams = 1500,
  String pdfFilename = 'lab-report.pdf',
  HoneyBatchCertificationModel? certification,
  HoneyBatchCertificationRequestModel? certificationRequest,
}) =>
    HoneyBatchModel(
      id: id,
      verificationToken: 'token-1',
      gatheringDate: DateTime(2025, 6, 1),
      amountGrams: amountGrams,
      processingMethod: processingMethod,
      honeyType: honeyType,
      pdfFilename: pdfFilename,
      pdfFileHash: 'hash-1',
      certification: certification,
      certificationRequest: certificationRequest,
      createdAt: DateTime(2025, 6, 1),
      updatedAt: DateTime(2025, 6, 1),
    );

void main() {
  late AppLocalizations l10n;
  late ApiClient apiClient;
  late MockHoneyBatchRepository repo;
  late HoneyBatchesCubit cubit;

  setUpAll(() async {
    l10n = await AppLocalizations.delegate.load(const Locale('en'));
  });

  setUp(() async {
    apiClient = await _fakeApiClient();
    repo = MockHoneyBatchRepository();
    cubit = HoneyBatchesCubit(repo: repo);
  });

  tearDown(() => cubit.close());

  // ApiClient is provided above MaterialApp (matching main.dart), so it stays
  // reachable from routes pushed onto the Navigator, not just the first page.
  Widget wrap(HoneyBatchModel batch) => RepositoryProvider<ApiClient>.value(
        value: apiClient,
        child: MaterialApp(
          localizationsDelegates: AppLocalizations.localizationsDelegates,
          supportedLocales: AppLocalizations.supportedLocales,
          locale: const Locale('en'),
          home: Scaffold(
            body: BlocProvider<HoneyBatchesCubit>.value(
              value: cubit,
              child: HoneyBatchCard(batch: batch),
            ),
          ),
        ),
      );

  group('HoneyBatchCard', () {
    testWidgets('shows honey type, processing method and amount in kg', (
      tester,
    ) async {
      await tester.pumpWidget(
        wrap(_batch(
          honeyType: 'Wildflower',
          processingMethod: ProcessingMethod.filtered,
          amountGrams: 2500,
        )),
      );

      expect(
        find.text(
            'Wildflower · ${processingMethodLabel(l10n, ProcessingMethod.filtered)}'),
        findsOneWidget,
      );
      expect(find.text('2.5 kg'), findsOneWidget);
    });

    testWidgets('shows "not certified" badge when certification is null', (
      tester,
    ) async {
      await tester.pumpWidget(wrap(_batch(certification: null)));
      expect(find.text(l10n.honeyBatchNotCertified), findsOneWidget);
    });

    testWidgets('shows the certification status badge when certified', (
      tester,
    ) async {
      await tester.pumpWidget(
        wrap(_batch(
          certification: const HoneyBatchCertificationModel(
            status: CertificationStatus.confirmed,
          ),
        )),
      );
      expect(find.text(l10n.honeyBatchStatusConfirmed), findsOneWidget);
    });

    testWidgets('shows "None" for pdf filename when it is empty', (
      tester,
    ) async {
      await tester.pumpWidget(wrap(_batch(pdfFilename: '')));
      expect(find.text(l10n.honeyBatchNoPdf), findsOneWidget);
    });

    testWidgets('shows the pdf filename when present', (tester) async {
      await tester.pumpWidget(wrap(_batch(pdfFilename: 'lab-report.pdf')));
      expect(find.text('lab-report.pdf'), findsOneWidget);
    });

    testWidgets('shows Edit menu item only when certification is null', (
      tester,
    ) async {
      await tester.pumpWidget(wrap(_batch(certification: null)));
      await tester.tap(find.byIcon(Icons.more_vert));
      await tester.pumpAndSettle();

      expect(find.text(l10n.generalEdit), findsOneWidget);
      expect(find.text(l10n.generalDelete), findsOneWidget);
    });

    testWidgets('tapping Edit opens the edit screen prefilled with the batch', (
      tester,
    ) async {
      final batch = _batch(certification: null);
      await tester.pumpWidget(wrap(batch));
      await tester.tap(find.byIcon(Icons.more_vert));
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.generalEdit));
      await tester.pumpAndSettle();

      expect(find.byType(CreateHoneyBatchScreen), findsOneWidget);
      expect(find.text(l10n.honeyBatchEditTitle), findsOneWidget);
      expect(find.text('Wildflower'), findsOneWidget);
    });

    testWidgets(
        'editing a batch with an attached pdf shows a clear button, and '
        'tapping it marks the pdf for removal', (tester) async {
      final batch = _batch(certification: null, pdfFilename: 'lab-report.pdf');
      await tester.pumpWidget(wrap(batch));
      await tester.tap(find.byIcon(Icons.more_vert));
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.generalEdit));
      await tester.pumpAndSettle();

      expect(find.text('lab-report.pdf'), findsOneWidget);
      expect(find.byIcon(Icons.close), findsOneWidget);

      await tester.tap(find.byIcon(Icons.close));
      await tester.pumpAndSettle();

      expect(find.text('lab-report.pdf'), findsNothing);
      expect(find.text(l10n.honeyBatchPdfLabel), findsOneWidget);
      expect(find.byIcon(Icons.close), findsNothing);
    });

    testWidgets(
        'submitting the edit form updates the card without throwing '
        '(regression: edit screen must not read an out-of-scope cubit)', (
      tester,
    ) async {
      apiClient = await _fakeApiClient(
        patchResponse: {
          'id': 1,
          'verification_token': 'token-1',
          'gathering_date': '2025-06-01T00:00:00Z',
          'amount_grams': 1500,
          'processing_method': 'raw',
          'honey_type': 'Acacia',
          'pdf_filename': 'lab-report.pdf',
          'pdf_file_hash': 'hash-1',
          'created_at': '2025-06-01T00:00:00Z',
          'updated_at': '2025-06-01T00:00:00Z',
          'certification': null,
        },
      );

      final original = _batch(certification: null);
      when(
        () => repo.listBatches(
            limit: any(named: 'limit'), offset: any(named: 'offset')),
      ).thenAnswer((_) async => (items: [original], total: 1));
      await cubit.load();

      await tester.pumpWidget(wrap(original));
      await tester.tap(find.byIcon(Icons.more_vert));
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.generalEdit));
      await tester.pumpAndSettle();

      await tester.enterText(
        find.widgetWithText(TextFormField, l10n.honeyBatchHoneyType),
        'Acacia',
      );
      await tester.tap(find.byIcon(Icons.check));
      await tester.pumpAndSettle();

      // No exception was thrown (a ProviderNotFoundException would fail this
      // test), the edit screen popped back, and the cubit's state (which the
      // real list screen renders from) reflects the update.
      expect(find.byType(CreateHoneyBatchScreen), findsNothing);
      final state = cubit.state as HoneyBatchesLoaded;
      expect(state.batches.single.honeyType, 'Acacia');
    });

    testWidgets('hides Edit menu item when batch is certified', (
      tester,
    ) async {
      await tester.pumpWidget(
        wrap(_batch(
          certification: const HoneyBatchCertificationModel(
            status: CertificationStatus.confirmed,
          ),
        )),
      );
      await tester.tap(find.byIcon(Icons.more_vert));
      await tester.pumpAndSettle();

      expect(find.text(l10n.generalEdit), findsNothing);
      expect(find.text(l10n.generalDelete), findsOneWidget);
    });

    testWidgets('tapping Certify calls cubit.requestCertification', (
      tester,
    ) async {
      when(() => repo.requestCertification(1)).thenAnswer(
        (_) async => _batch(
          certification:
              const HoneyBatchCertificationModel(status: CertificationStatus.queued),
        ),
      );

      await tester.pumpWidget(wrap(_batch(certification: null)));
      await tester.tap(find.byIcon(Icons.more_vert));
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.honeyBatchCertify));
      await tester.pumpAndSettle();
      await _solvePuzzle(tester, l10n.generalConfirm);
      await tester.pump();

      verify(() => repo.requestCertification(1)).called(1);
    });

    testWidgets('does not certify when the confirmation dialog is cancelled', (
      tester,
    ) async {
      await tester.pumpWidget(wrap(_batch(certification: null)));
      await tester.tap(find.byIcon(Icons.more_vert));
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.honeyBatchCertify));
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.generalCancel));
      await tester.pump();

      verifyNever(() => repo.requestCertification(1));
    });

    testWidgets(
        'shows the pending certification request badge and hides the '
        'Certify button', (tester) async {
      await tester.pumpWidget(
        wrap(_batch(
          certification: null,
          certificationRequest: const HoneyBatchCertificationRequestModel(
            status: CertificationRequestStatus.pending,
          ),
        )),
      );

      expect(find.text(l10n.honeyBatchCertRequestStatusPending), findsOneWidget);
      expect(find.text(l10n.honeyBatchNotCertified), findsNothing);
      expect(find.text(l10n.honeyBatchCertify), findsNothing);
    });

    testWidgets(
        'shows the rejected certification request badge with the rejection '
        'reason as a tooltip, and still allows re-requesting certification', (
      tester,
    ) async {
      await tester.pumpWidget(
        wrap(_batch(
          certification: null,
          certificationRequest: const HoneyBatchCertificationRequestModel(
            status: CertificationRequestStatus.rejected,
            rejectionReason: 'Missing lab report',
          ),
        )),
      );

      expect(find.text(l10n.honeyBatchCertRequestStatusRejected), findsOneWidget);
      expect(
        find.byWidgetPredicate(
          (w) => w is Tooltip && w.message == 'Missing lab report',
        ),
        findsOneWidget,
      );
      await tester.tap(find.byIcon(Icons.more_vert));
      await tester.pumpAndSettle();
      expect(find.text(l10n.honeyBatchCertify), findsOneWidget);

      when(() => repo.requestCertification(1)).thenAnswer(
        (_) async => _batch(
          certification:
              const HoneyBatchCertificationModel(status: CertificationStatus.queued),
        ),
      );
      await tester.tap(find.text(l10n.honeyBatchCertify));
      await tester.pumpAndSettle();
      await _solvePuzzle(tester, l10n.generalConfirm);
      await tester.pump();

      verify(() => repo.requestCertification(1)).called(1);
    });

    testWidgets('shows a view icon next to the pdf filename when present', (
      tester,
    ) async {
      await tester.pumpWidget(wrap(_batch(pdfFilename: 'lab-report.pdf')));
      expect(find.byIcon(Icons.visibility_outlined), findsOneWidget);
    });

    testWidgets('shows no view icon when there is no pdf', (tester) async {
      await tester.pumpWidget(wrap(_batch(pdfFilename: '')));
      expect(find.byIcon(Icons.visibility_outlined), findsNothing);
    });

    testWidgets('shows View/Download QR buttons when confirmed', (
      tester,
    ) async {
      await tester.pumpWidget(
        wrap(_batch(
          certification: const HoneyBatchCertificationModel(
            status: CertificationStatus.confirmed,
          ),
        )),
      );

      expect(find.text(l10n.honeyBatchViewQr), findsOneWidget);
      expect(find.text(l10n.honeyBatchDownloadQr), findsOneWidget);
    });
  });

  group('HoneyBatchCard in a list', () {
    // ApiClient is provided above MaterialApp, same as wrap() above.
    Widget wrapList() => RepositoryProvider<ApiClient>.value(
          value: apiClient,
          child: MaterialApp(
            localizationsDelegates: AppLocalizations.localizationsDelegates,
            supportedLocales: AppLocalizations.supportedLocales,
            locale: const Locale('en'),
            home: Scaffold(
              body: BlocProvider<HoneyBatchesCubit>.value(
                value: cubit,
                child: BlocBuilder<HoneyBatchesCubit, HoneyBatchesState>(
                  builder: (context, state) {
                    final batches = (state as HoneyBatchesLoaded).batches;
                    return Column(
                      children: [
                        for (final batch in batches)
                          HoneyBatchCard(key: ValueKey(batch.id), batch: batch),
                      ],
                    );
                  },
                ),
              ),
            ),
          ),
        );

    testWidgets(
        'regression: deleting one card does not leave the card that shifts '
        'into its list position stuck showing a spinner', (tester) async {
      final batchA = _batch(id: 1, honeyType: 'Wildflower');
      final batchB = _batch(id: 2, honeyType: 'Acacia');

      when(
        () => repo.listBatches(
            limit: any(named: 'limit'), offset: any(named: 'offset')),
      ).thenAnswer((_) async => (items: [batchA, batchB], total: 2));
      await cubit.load();

      when(() => repo.deleteBatch(1)).thenAnswer((_) async {});
      when(
        () => repo.listBatches(
            limit: any(named: 'limit'), offset: any(named: 'offset')),
      ).thenAnswer((_) async => (items: [batchB], total: 1));

      await tester.pumpWidget(wrapList());
      expect(find.byIcon(Icons.more_vert), findsNWidgets(2));

      await tester.tap(find.byIcon(Icons.more_vert).first);
      await tester.pumpAndSettle();
      await tester.tap(find.text(l10n.generalDelete));
      await tester.pumpAndSettle();
      await _solvePuzzle(tester, l10n.generalDelete);

      // Only batch B remains, and it must still show its own menu button —
      // not a spinner inherited from batch A's now-removed card.
      expect(find.textContaining('Acacia'), findsOneWidget);
      expect(find.textContaining('Wildflower'), findsNothing);
      expect(find.byIcon(Icons.more_vert), findsOneWidget);
      expect(find.byType(CircularProgressIndicator), findsNothing);
    });
  });
}
