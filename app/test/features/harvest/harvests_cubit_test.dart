import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:app/features/harvest/cubit/harvests_cubit.dart';
import 'package:app/features/harvest/data/harvest_model.dart';
import 'package:app/features/harvest/data/harvest_repository.dart';

class MockHarvestRepository extends Mock implements HarvestRepository {}

Harvest _fakeHarvest({int id = 1}) => Harvest(
      id: id,
      hiveId: 10,
      harvestedAt: DateTime(2026, 6, 1),
      frames: 5,
      halfFrames: 2,
      kilograms: 12.50,
      notes: '',
      harvestedByName: null,
    );

void main() {
  late MockHarvestRepository repo;
  late HarvestsCubit cubit;

  setUp(() {
    repo = MockHarvestRepository();
    cubit = HarvestsCubit(repo: repo, apiaryId: 1, hiveId: 10);
  });

  tearDown(() => cubit.close());

  group('Harvest.fromJson', () {
    test('parses all fields correctly', () {
      final json = {
        'id': 5,
        'hive_id': 10,
        'harvested_at': '2026-06-01T00:00:00Z',
        'harvested_by_name': 'Jan Nowak',
        'frames': 8,
        'half_frames': 3,
        'kilograms': 24.75,
      };
      final h = Harvest.fromJson(json);
      expect(h.id, 5);
      expect(h.hiveId, 10);
      expect(h.harvestedByName, 'Jan Nowak');
      expect(h.frames, 8);
      expect(h.halfFrames, 3);
      expect(h.kilograms, 24.75);
    });

    test('handles missing optional fields with defaults', () {
      final json = {
        'id': 1,
        'hive_id': 10,
        'harvested_at': '2026-06-01T00:00:00Z',
      };
      final h = Harvest.fromJson(json);
      expect(h.frames, 0);
      expect(h.halfFrames, 0);
      expect(h.kilograms, 0.0);
    });
  });

  ({List<Harvest> items, int total}) _result(List<Harvest> items) =>
      (items: items, total: items.length);

  group('load', () {
    final harvests = [_fakeHarvest(id: 1), _fakeHarvest(id: 2)];

    blocTest<HarvestsCubit, HarvestsState>(
      'emits [HarvestsLoading, HarvestsLoaded] on success',
      build: () {
        when(
          () => repo.listHarvests(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => _result(harvests));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<HarvestsLoading>(),
        isA<HarvestsLoaded>()
            .having((s) => s.harvests.length, 'length', 2),
      ],
    );

    blocTest<HarvestsCubit, HarvestsState>(
      'emits [HarvestsLoading, HarvestsError] on failure',
      build: () {
        when(
          () => repo.listHarvests(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenThrow(Exception('network error'));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [isA<HarvestsLoading>(), isA<HarvestsError>()],
    );

    blocTest<HarvestsCubit, HarvestsState>(
      'totalPages is at least 1 when total is 0',
      build: () {
        when(
          () => repo.listHarvests(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => (items: <Harvest>[], total: 0));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<HarvestsLoading>(),
        isA<HarvestsLoaded>()
            .having((s) => s.totalPages, 'totalPages', 1),
      ],
    );
  });

  group('goToPage', () {
    blocTest<HarvestsCubit, HarvestsState>(
      'emits correct currentPage and totalPages',
      build: () {
        when(
          () => repo.listHarvests(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => (items: [_fakeHarvest()], total: 15));
        return cubit;
      },
      act: (c) => c.goToPage(2),
      expect: () => [
        isA<HarvestsLoading>(),
        isA<HarvestsLoaded>()
            .having((s) => s.currentPage, 'currentPage', 2)
            .having((s) => s.totalPages, 'totalPages', 2),
      ],
    );
  });

  group('delete', () {
    blocTest<HarvestsCubit, HarvestsState>(
      'emits [HarvestsLoading, HarvestsLoaded] on success',
      build: () {
        when(
          () => repo.deleteHarvest(apiaryId: 1, hiveId: 10, harvestId: 1),
        ).thenAnswer((_) async {});
        when(
          () => repo.listHarvests(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => _result([]));
        return cubit;
      },
      act: (c) => c.delete(1),
      expect: () => [
        isA<HarvestsLoading>(),
        isA<HarvestsLoaded>()
            .having((s) => s.harvests, 'harvests', isEmpty),
      ],
    );

    blocTest<HarvestsCubit, HarvestsState>(
      'backs up to previous page when current page becomes empty after delete',
      build: () {
        when(
          () => repo.deleteHarvest(apiaryId: 1, hiveId: 10, harvestId: 1),
        ).thenAnswer((_) async {});
        var call = 0;
        when(
          () => repo.listHarvests(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async {
          call++;
          if (call == 1) return (items: <Harvest>[], total: 1);
          return (items: [_fakeHarvest()], total: 1);
        });
        return cubit;
      },
      seed: () => HarvestsLoaded([], currentPage: 2, totalPages: 2),
      act: (c) => c.delete(1),
      expect: () => [
        isA<HarvestsLoading>(),
        isA<HarvestsLoading>(),
        isA<HarvestsLoaded>()
            .having((s) => s.currentPage, 'currentPage', 1)
            .having((s) => s.harvests, 'harvests', isNotEmpty),
      ],
    );

    blocTest<HarvestsCubit, HarvestsState>(
      'emits [HarvestsLoading, HarvestsError] on failure',
      build: () {
        when(
          () => repo.deleteHarvest(apiaryId: 1, hiveId: 10, harvestId: 1),
        ).thenThrow(Exception('error'));
        return cubit;
      },
      act: (c) => c.delete(1),
      expect: () => [isA<HarvestsLoading>(), isA<HarvestsError>()],
    );
  });
}
