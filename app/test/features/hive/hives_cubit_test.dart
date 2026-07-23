import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:app/features/hive/cubit/hives_cubit.dart';
import 'package:app/features/hive/data/hive_model.dart';
import 'package:app/features/hive/data/hive_repository.dart';

class MockHiveRepository extends Mock implements HiveRepository {}

void main() {
  late MockHiveRepository repo;
  late HivesCubit cubit;

  setUp(() {
    repo = MockHiveRepository();
    cubit = HivesCubit(repo: repo, apiaryId: 1);
  });

  tearDown(() => cubit.close());

  group('Hive.fromJson', () {
    test('parses all fields correctly', () {
      final hive = Hive.fromJson({
        'id': 1,
        'apiary_id': 2,
        'name': 'Ul 1',
        'type': 'langstroth',
        'active': true,
        'queenless': true,
        'ready_for_harvest': true,
        'grid_row': 0,
        'grid_col': 1,
        'diseases': [
          {'id': 10, 'disease': 'varroa', 'created_at': '2026-06-01T00:00:00Z'},
        ],
        'last_inspected_at': '2026-06-01T10:00:00Z',
      });
      expect(hive.id, 1);
      expect(hive.apiaryId, 2);
      expect(hive.name, 'Ul 1');
      expect(hive.type, 'langstroth');
      expect(hive.active, true);
      expect(hive.queenless, true);
      expect(hive.readyForHarvest, true);
      expect(hive.gridRow, 0);
      expect(hive.gridCol, 1);
      expect(hive.diseases.length, 1);
      expect(hive.diseases.first.disease, 'varroa');
      expect(hive.lastInspectedAt, DateTime.utc(2026, 6, 1, 10));
    });

    test('parses inactive hive with defaults', () {
      final hive = Hive.fromJson({
        'id': 5,
        'apiary_id': 1,
        'name': 'Old hive',
        'type': 'dadant',
        'active': false,
        'grid_row': 2,
        'grid_col': 3,
      });
      expect(hive.active, false);
      expect(hive.type, 'dadant');
      expect(hive.queenless, false);
      expect(hive.readyForHarvest, false);
      expect(hive.diseases, isEmpty);
      expect(hive.lastInspectedAt, isNull);
    });
  });

  group('load', () {
    final hives = [
      const Hive(
        id: 1,
        apiaryId: 1,
        name: 'Ul 1',
        type: 'langstroth',
        active: true,
        queenless: false,
        readyForHarvest: false,
        needsFood: false,
        gridRow: 0,
        gridCol: 0,
      ),
    ];

    blocTest<HivesCubit, HivesState>(
      'emits [HivesLoading, HivesLoaded] on success',
      build: () {
        when(() => repo.listHives(1)).thenAnswer((_) async => hives);
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<HivesLoading>(),
        isA<HivesLoaded>().having((s) => s.hives.length, 'length', 1),
      ],
    );

    blocTest<HivesCubit, HivesState>(
      'emits [HivesLoading, HivesLoaded] with empty list',
      build: () {
        when(() => repo.listHives(1)).thenAnswer((_) async => []);
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<HivesLoading>(),
        isA<HivesLoaded>().having((s) => s.hives, 'hives', isEmpty),
      ],
    );

    blocTest<HivesCubit, HivesState>(
      'emits [HivesLoading, HivesError] on failure',
      build: () {
        when(() => repo.listHives(1)).thenThrow(Exception('network error'));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [isA<HivesLoading>(), isA<HivesError>()],
    );
  });

  group('delete', () {
    blocTest<HivesCubit, HivesState>(
      'emits [HivesLoading, HivesLoaded] on success',
      build: () {
        when(() => repo.deleteHive(apiaryId: 1, hiveId: 1))
            .thenAnswer((_) async {});
        when(() => repo.listHives(1)).thenAnswer((_) async => []);
        return cubit;
      },
      act: (c) => c.delete(1),
      expect: () => [
        isA<HivesLoading>(),
        isA<HivesLoaded>().having((s) => s.hives, 'hives', isEmpty),
      ],
    );

    blocTest<HivesCubit, HivesState>(
      'emits [HivesLoading, HivesError] on failure',
      build: () {
        when(() => repo.deleteHive(apiaryId: 1, hiveId: 1))
            .thenThrow(Exception('error'));
        return cubit;
      },
      act: (c) => c.delete(1),
      expect: () => [isA<HivesLoading>(), isA<HivesError>()],
    );
  });

  group('move', () {
    const hive = Hive(
      id: 1,
      apiaryId: 1,
      name: 'Ul 1',
      type: 'langstroth',
      active: true,
      queenless: false,
      readyForHarvest: false,
      needsFood: false,
      gridRow: 0,
      gridCol: 0,
    );

    blocTest<HivesCubit, HivesState>(
      'optimistically updates position then confirms via API',
      build: () {
        when(() => repo.moveHive(apiaryId: 1, hiveId: 1, row: 1, col: 2))
            .thenAnswer((_) async {});
        return cubit;
      },
      seed: () => HivesLoaded([hive]),
      act: (c) => c.move(1, 1, 2),
      expect: () => [
        isA<HivesLoaded>().having(
          (s) => (s.hives.first.gridRow, s.hives.first.gridCol),
          'position',
          (1, 2),
        ),
      ],
    );

    blocTest<HivesCubit, HivesState>(
      'reverts to server state on API failure',
      build: () {
        when(() => repo.moveHive(apiaryId: 1, hiveId: 1, row: 1, col: 2))
            .thenThrow(Exception('error'));
        when(() => repo.listHives(1)).thenAnswer((_) async => [hive]);
        return cubit;
      },
      seed: () => HivesLoaded([hive]),
      act: (c) => c.move(1, 1, 2),
      expect: () => [
        isA<HivesLoaded>().having(
          (s) => (s.hives.first.gridRow, s.hives.first.gridCol),
          'optimistic position',
          (1, 2),
        ),
        isA<HivesLoading>(),
        isA<HivesLoaded>().having(
          (s) => (s.hives.first.gridRow, s.hives.first.gridCol),
          'reverted position',
          (0, 0),
        ),
      ],
    );

    blocTest<HivesCubit, HivesState>(
      'does nothing when not in loaded state',
      build: () => cubit,
      act: (c) => c.move(1, 1, 2),
      expect: () => [],
    );
  });

  group('hive list sort order', () {
    // Mirrors the comparator used in _HiveListDialog:
    // null (no inspection) first, then oldest to newest.
    List<Hive> sortHives(List<Hive> hives) => [...hives]..sort((a, b) {
        if (a.lastInspectedAt == null && b.lastInspectedAt == null) return 0;
        if (a.lastInspectedAt == null) return -1;
        if (b.lastInspectedAt == null) return 1;
        return a.lastInspectedAt!.compareTo(b.lastInspectedAt!);
      });

    Hive hiveWithDate(int id, DateTime? date) => Hive(
          id: id,
          apiaryId: 1,
          name: 'Ul $id',
          type: 'langstroth',
          active: true,
          queenless: false,
          readyForHarvest: false,
          needsFood: false,
          gridRow: 0,
          gridCol: id,
          lastInspectedAt: date,
        );

    test('no-inspection hive sorts before inspected hives', () {
      final hives = [
        hiveWithDate(1, DateTime(2026, 6, 6)),
        hiveWithDate(2, null),
      ];
      final sorted = sortHives(hives);
      expect(sorted[0].lastInspectedAt, isNull);
      expect(sorted[1].lastInspectedAt, DateTime(2026, 6, 6));
    });

    test('oldest inspection sorts before most recent', () {
      final hives = [
        hiveWithDate(1, DateTime(2026, 6, 6)),
        hiveWithDate(2, DateTime(2026, 6, 1)),
        hiveWithDate(3, DateTime(2026, 5, 15)),
      ];
      final sorted = sortHives(hives);
      expect(sorted[0].lastInspectedAt, DateTime(2026, 5, 15));
      expect(sorted[1].lastInspectedAt, DateTime(2026, 6, 1));
      expect(sorted[2].lastInspectedAt, DateTime(2026, 6, 6));
    });

    test('multiple no-inspection hives preserve relative order', () {
      final hives = [
        hiveWithDate(1, null),
        hiveWithDate(2, DateTime(2026, 6, 1)),
        hiveWithDate(3, null),
      ];
      final sorted = sortHives(hives);
      expect(sorted[0].lastInspectedAt, isNull);
      expect(sorted[1].lastInspectedAt, isNull);
      expect(sorted[2].lastInspectedAt, DateTime(2026, 6, 1));
    });

    test('all inspected sorts oldest to newest', () {
      final hives = [
        hiveWithDate(1, DateTime(2026, 6, 5)),
        hiveWithDate(2, DateTime(2026, 6, 1)),
        hiveWithDate(3, DateTime(2026, 6, 10)),
      ];
      final sorted = sortHives(hives);
      expect(sorted.map((h) => h.id).toList(), [2, 1, 3]);
    });

    test('all no-inspection returns same count', () {
      final hives = [hiveWithDate(1, null), hiveWithDate(2, null)];
      final sorted = sortHives(hives);
      expect(sorted.length, 2);
      expect(sorted.every((h) => h.lastInspectedAt == null), isTrue);
    });
  });
}
