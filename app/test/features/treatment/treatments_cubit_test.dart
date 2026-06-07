import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:app/features/treatment/cubit/treatments_cubit.dart';
import 'package:app/features/treatment/data/treatment_model.dart';
import 'package:app/features/treatment/data/treatment_repository.dart';

class MockTreatmentRepository extends Mock implements TreatmentRepository {}

Treatment _fakeTreatment({int id = 1}) => Treatment(
      id: id,
      hiveId: 10,
      treatedAt: DateTime(2025, 6, 1),
      medicineName: 'Apiwarol',
      dose: '2 strips',
      notes: '',
    );

void main() {
  late MockTreatmentRepository repo;
  late TreatmentsCubit cubit;

  setUp(() {
    repo = MockTreatmentRepository();
    cubit = TreatmentsCubit(repo: repo, apiaryId: 1, hiveId: 10);
  });

  tearDown(() => cubit.close());

  group('Treatment.fromJson', () {
    test('parses all fields correctly', () {
      final json = {
        'id': 5,
        'hive_id': 10,
        'treated_at': '2025-06-01T10:00:00Z',
        'medicine_name': 'Apiwarol',
        'dose': '2 strips',
        'notes': 'Applied on frames',
        'treated_by_name': 'Alice',
      };
      final t = Treatment.fromJson(json);
      expect(t.id, 5);
      expect(t.hiveId, 10);
      expect(t.medicineName, 'Apiwarol');
      expect(t.dose, '2 strips');
      expect(t.notes, 'Applied on frames');
      expect(t.treatedByName, 'Alice');
    });

    test('handles missing optional fields', () {
      final json = {
        'id': 1,
        'hive_id': 10,
        'treated_at': '2025-06-01T10:00:00Z',
        'medicine_name': 'Apiwarol',
      };
      final t = Treatment.fromJson(json);
      expect(t.dose, '1');
      expect(t.notes, '');
      expect(t.treatedByName, isNull);
    });
  });

  ({List<Treatment> items, int total}) _result(List<Treatment> items) =>
      (items: items, total: items.length);

  group('load', () {
    final treatments = [_fakeTreatment(id: 1), _fakeTreatment(id: 2)];

    blocTest<TreatmentsCubit, TreatmentsState>(
      'emits [TreatmentsLoading, TreatmentsLoaded] on success',
      build: () {
        when(
          () => repo.listTreatments(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => _result(treatments));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<TreatmentsLoading>(),
        isA<TreatmentsLoaded>()
            .having((s) => s.treatments.length, 'length', 2),
      ],
    );

    blocTest<TreatmentsCubit, TreatmentsState>(
      'emits [TreatmentsLoading, TreatmentsError] on failure',
      build: () {
        when(
          () => repo.listTreatments(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenThrow(Exception('network error'));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [isA<TreatmentsLoading>(), isA<TreatmentsError>()],
    );

    blocTest<TreatmentsCubit, TreatmentsState>(
      'totalPages is at least 1 when total is 0',
      build: () {
        when(
          () => repo.listTreatments(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => (items: <Treatment>[], total: 0));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<TreatmentsLoading>(),
        isA<TreatmentsLoaded>()
            .having((s) => s.totalPages, 'totalPages', 1),
      ],
    );
  });

  group('goToPage', () {
    blocTest<TreatmentsCubit, TreatmentsState>(
      'emits correct currentPage and totalPages',
      build: () {
        when(
          () => repo.listTreatments(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => (items: [_fakeTreatment()], total: 15));
        return cubit;
      },
      act: (c) => c.goToPage(2),
      expect: () => [
        isA<TreatmentsLoading>(),
        isA<TreatmentsLoaded>()
            .having((s) => s.currentPage, 'currentPage', 2)
            .having((s) => s.totalPages, 'totalPages', 2),
      ],
    );
  });

  group('delete', () {
    blocTest<TreatmentsCubit, TreatmentsState>(
      'emits [TreatmentsLoading, TreatmentsLoaded] on success',
      build: () {
        when(
          () => repo.deleteTreatment(apiaryId: 1, hiveId: 10, treatmentId: 1),
        ).thenAnswer((_) async {});
        when(
          () => repo.listTreatments(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => _result([]));
        return cubit;
      },
      act: (c) => c.delete(1),
      expect: () => [
        isA<TreatmentsLoading>(),
        isA<TreatmentsLoaded>()
            .having((s) => s.treatments, 'treatments', isEmpty),
      ],
    );

    blocTest<TreatmentsCubit, TreatmentsState>(
      'backs up to previous page when current page becomes empty after delete',
      build: () {
        when(
          () => repo.deleteTreatment(apiaryId: 1, hiveId: 10, treatmentId: 1),
        ).thenAnswer((_) async {});
        var call = 0;
        when(
          () => repo.listTreatments(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async {
          call++;
          if (call == 1) return (items: <Treatment>[], total: 1);
          return (items: [_fakeTreatment()], total: 1);
        });
        return cubit;
      },
      seed: () => TreatmentsLoaded([], currentPage: 2, totalPages: 2),
      act: (c) => c.delete(1),
      expect: () => [
        isA<TreatmentsLoading>(),
        isA<TreatmentsLoading>(),
        isA<TreatmentsLoaded>()
            .having((s) => s.currentPage, 'currentPage', 1)
            .having((s) => s.treatments, 'treatments', isNotEmpty),
      ],
    );

    blocTest<TreatmentsCubit, TreatmentsState>(
      'emits [TreatmentsLoading, TreatmentsError] on failure',
      build: () {
        when(
          () => repo.deleteTreatment(apiaryId: 1, hiveId: 10, treatmentId: 1),
        ).thenThrow(Exception('error'));
        return cubit;
      },
      act: (c) => c.delete(1),
      expect: () => [isA<TreatmentsLoading>(), isA<TreatmentsError>()],
    );
  });
}
