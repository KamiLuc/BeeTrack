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
        'grid_row': 0,
        'grid_col': 1,
      });
      expect(hive.id, 1);
      expect(hive.apiaryId, 2);
      expect(hive.name, 'Ul 1');
      expect(hive.type, 'langstroth');
      expect(hive.active, true);
      expect(hive.gridRow, 0);
      expect(hive.gridCol, 1);
    });

    test('parses inactive hive', () {
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
}
