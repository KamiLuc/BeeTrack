import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:app/features/honey_batch/cubit/honey_batches_cubit.dart';
import 'package:app/features/honey_batch/data/honey_batch_model.dart';
import 'package:app/features/honey_batch/data/honey_batch_repository.dart';
import 'package:app/features/honey_batch/data/processing_method.dart';

class MockHoneyBatchRepository extends Mock implements HoneyBatchRepository {}

HoneyBatchModel _fakeBatch({int id = 1}) => HoneyBatchModel(
      id: id,
      verificationToken: 'token-$id',
      gatheringDate: DateTime(2025, 6, 1),
      amountGrams: 1000,
      processingMethod: ProcessingMethod.raw,
      honeyType: 'Wildflower',
      pdfFilename: 'lab-$id.pdf',
      pdfFileHash: 'hash-$id',
      createdAt: DateTime(2025, 6, 1),
      updatedAt: DateTime(2025, 6, 1),
    );

void main() {
  late MockHoneyBatchRepository repo;
  late HoneyBatchesCubit cubit;

  setUpAll(() {
    registerFallbackValue(ProcessingMethod.raw);
  });

  setUp(() {
    repo = MockHoneyBatchRepository();
    cubit = HoneyBatchesCubit(repo: repo);
  });

  tearDown(() => cubit.close());

  ({List<HoneyBatchModel> items, int total}) _result(
          List<HoneyBatchModel> items) =>
      (items: items, total: items.length);

  group('load', () {
    final batches = [_fakeBatch(id: 1), _fakeBatch(id: 2)];

    blocTest<HoneyBatchesCubit, HoneyBatchesState>(
      'emits [HoneyBatchesLoading, HoneyBatchesLoaded] on success',
      build: () {
        when(
          () => repo.listBatches(
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => _result(batches));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<HoneyBatchesLoading>(),
        isA<HoneyBatchesLoaded>()
            .having((s) => s.batches.length, 'length', 2),
      ],
    );

    blocTest<HoneyBatchesCubit, HoneyBatchesState>(
      'emits [HoneyBatchesLoading, HoneyBatchesError] on failure',
      build: () {
        when(
          () => repo.listBatches(
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenThrow(Exception('network error'));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [isA<HoneyBatchesLoading>(), isA<HoneyBatchesError>()],
    );
  });

  group('goToPage', () {
    blocTest<HoneyBatchesCubit, HoneyBatchesState>(
      'emits correct currentPage and totalPages',
      build: () {
        when(
          () => repo.listBatches(
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => (items: [_fakeBatch()], total: 25));
        return cubit;
      },
      act: (c) => c.goToPage(2),
      expect: () => [
        isA<HoneyBatchesLoading>(),
        isA<HoneyBatchesLoaded>()
            .having((s) => s.currentPage, 'currentPage', 2)
            .having((s) => s.totalPages, 'totalPages', 2),
      ],
      verify: (_) {
        verify(() => repo.listBatches(limit: 20, offset: 20)).called(1);
      },
    );
  });

  group('create', () {
    test('calls repo.createBatch with correct args and reloads page 1',
        () async {
      final created = _fakeBatch(id: 3);
      when(
        () => repo.createBatch(
          gatheringDate: any(named: 'gatheringDate'),
          amountGrams: any(named: 'amountGrams'),
          processingMethod: any(named: 'processingMethod'),
          honeyType: any(named: 'honeyType'),
          pdfBytes: any(named: 'pdfBytes'),
          pdfFilename: any(named: 'pdfFilename'),
          requestCertification: any(named: 'requestCertification'),
        ),
      ).thenAnswer((_) async => created);
      when(
        () => repo.listBatches(
            limit: any(named: 'limit'), offset: any(named: 'offset')),
      ).thenAnswer((_) async => _result([created]));

      final gatheringDate = DateTime(2025, 6, 1);
      final result = await cubit.create(
        gatheringDate: gatheringDate,
        amountGrams: 1000,
        processingMethod: ProcessingMethod.filtered,
        honeyType: 'Wildflower',
        pdfBytes: [1, 2, 3],
        pdfFilename: 'lab.pdf',
        requestCertification: true,
      );

      expect(result, created);
      verify(
        () => repo.createBatch(
          gatheringDate: gatheringDate,
          amountGrams: 1000,
          processingMethod: ProcessingMethod.filtered,
          honeyType: 'Wildflower',
          pdfBytes: [1, 2, 3],
          pdfFilename: 'lab.pdf',
          requestCertification: true,
        ),
      ).called(1);
      verify(() => repo.listBatches(limit: 20, offset: 0)).called(1);
      expect(cubit.state, isA<HoneyBatchesLoaded>());
    });

    test('rethrows on repo failure', () async {
      when(
        () => repo.createBatch(
          gatheringDate: any(named: 'gatheringDate'),
          amountGrams: any(named: 'amountGrams'),
          processingMethod: any(named: 'processingMethod'),
          honeyType: any(named: 'honeyType'),
          pdfBytes: any(named: 'pdfBytes'),
          pdfFilename: any(named: 'pdfFilename'),
          requestCertification: any(named: 'requestCertification'),
        ),
      ).thenThrow(Exception('upload failed'));

      expect(
        () => cubit.create(
          gatheringDate: DateTime(2025, 6, 1),
          amountGrams: 1000,
          processingMethod: ProcessingMethod.raw,
          honeyType: 'Wildflower',
          pdfBytes: [1, 2, 3],
          pdfFilename: 'lab.pdf',
        ),
        throwsException,
      );
    });
  });

  group('requestCertification', () {
    blocTest<HoneyBatchesCubit, HoneyBatchesState>(
      'replaces the matching batch in a Loaded state',
      build: () {
        final updated = _fakeBatch(id: 1);
        when(() => repo.requestCertification(1))
            .thenAnswer((_) async => updated);
        return cubit;
      },
      seed: () => HoneyBatchesLoaded(
        [_fakeBatch(id: 1), _fakeBatch(id: 2)],
        currentPage: 1,
        totalPages: 1,
      ),
      act: (c) => c.requestCertification(1),
      expect: () => [
        isA<HoneyBatchesLoaded>()
            .having((s) => s.batches.length, 'length', 2)
            .having((s) => s.batches.first.id, 'first id', 1),
      ],
    );

    test('no-ops when state is not Loaded', () async {
      when(() => repo.requestCertification(1))
          .thenAnswer((_) async => _fakeBatch(id: 1));

      await cubit.requestCertification(1);

      expect(cubit.state, isA<HoneyBatchesInitial>());
    });

    test('rethrows on repo failure without touching state', () async {
      when(() => repo.requestCertification(1))
          .thenThrow(Exception('certification failed'));

      expect(() => cubit.requestCertification(1), throwsException);
    });
  });

  group('replaceInList', () {
    blocTest<HoneyBatchesCubit, HoneyBatchesState>(
      'replaces the matching batch in a Loaded state',
      build: () => cubit,
      seed: () => HoneyBatchesLoaded(
        [_fakeBatch(id: 1), _fakeBatch(id: 2)],
        currentPage: 1,
        totalPages: 1,
      ),
      act: (c) => c.replaceInList(_fakeBatch(id: 1)),
      expect: () => [
        isA<HoneyBatchesLoaded>()
            .having((s) => s.batches.length, 'length', 2)
            .having((s) => s.batches.first.id, 'first id', 1),
      ],
    );

    test('no-ops when state is not Loaded', () {
      cubit.replaceInList(_fakeBatch(id: 1));

      expect(cubit.state, isA<HoneyBatchesInitial>());
    });
  });

  group('delete', () {
    blocTest<HoneyBatchesCubit, HoneyBatchesState>(
      'reloads the current page on success',
      build: () {
        when(() => repo.deleteBatch(1)).thenAnswer((_) async {});
        when(
          () => repo.listBatches(
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => _result([_fakeBatch(id: 2)]));
        return cubit;
      },
      seed: () => HoneyBatchesLoaded(
        [_fakeBatch(id: 1), _fakeBatch(id: 2)],
        currentPage: 1,
        totalPages: 1,
      ),
      act: (c) => c.delete(1),
      expect: () => [
        isA<HoneyBatchesLoaded>()
            .having((s) => s.batches.length, 'length', 1)
            .having((s) => s.currentPage, 'currentPage', 1),
      ],
    );

    blocTest<HoneyBatchesCubit, HoneyBatchesState>(
      'backs up to previous page when current page becomes empty after delete',
      build: () {
        when(() => repo.deleteBatch(1)).thenAnswer((_) async {});
        var call = 0;
        when(
          () => repo.listBatches(
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async {
          call++;
          if (call == 1) return _result([]);
          return _result([_fakeBatch(id: 2)]);
        });
        return cubit;
      },
      seed: () => HoneyBatchesLoaded([_fakeBatch(id: 1)],
          currentPage: 2, totalPages: 2),
      act: (c) => c.delete(1),
      expect: () => [
        isA<HoneyBatchesLoading>(),
        isA<HoneyBatchesLoaded>()
            .having((s) => s.currentPage, 'currentPage', 1)
            .having((s) => s.batches, 'batches', isNotEmpty),
      ],
    );

    test('no-ops when state is not Loaded', () async {
      await cubit.delete(1);

      expect(cubit.state, isA<HoneyBatchesInitial>());
      verifyNever(() => repo.deleteBatch(any()));
    });

    test('rethrows on repo failure without touching state', () async {
      when(() => repo.deleteBatch(1)).thenThrow(Exception('delete failed'));

      cubit.emit(HoneyBatchesLoaded(
        [_fakeBatch(id: 1)],
        currentPage: 1,
        totalPages: 1,
      ));

      expect(() => cubit.delete(1), throwsException);
    });
  });
}
