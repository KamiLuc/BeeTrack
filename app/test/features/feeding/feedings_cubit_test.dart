import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:app/features/feeding/cubit/feedings_cubit.dart';
import 'package:app/features/feeding/data/feeding_model.dart';
import 'package:app/features/feeding/data/feeding_repository.dart';

class MockFeedingRepository extends Mock implements FeedingRepository {}

Feeding _fakeFeeding({int id = 1}) => Feeding(
      id: id,
      hiveId: 10,
      fedAt: DateTime(2025, 6, 1),
      feedType: 'Sugar syrup',
      amount: '1L',
      notes: '',
    );

void main() {
  late MockFeedingRepository repo;
  late FeedingsCubit cubit;

  setUp(() {
    repo = MockFeedingRepository();
    cubit = FeedingsCubit(repo: repo, apiaryId: 1, hiveId: 10);
  });

  tearDown(() => cubit.close());

  group('Feeding.fromJson', () {
    test('parses all fields correctly', () {
      final json = {
        'id': 5,
        'hive_id': 10,
        'fed_at': '2025-06-01T10:00:00Z',
        'feed_type': 'Sugar syrup',
        'amount': '1L',
        'notes': 'Applied on frames',
        'fed_by_name': 'Alice',
      };
      final f = Feeding.fromJson(json);
      expect(f.id, 5);
      expect(f.hiveId, 10);
      expect(f.feedType, 'Sugar syrup');
      expect(f.amount, '1L');
      expect(f.notes, 'Applied on frames');
      expect(f.fedByName, 'Alice');
    });

    test('handles missing optional fields', () {
      final json = {
        'id': 1,
        'hive_id': 10,
        'fed_at': '2025-06-01T10:00:00Z',
      };
      final f = Feeding.fromJson(json);
      expect(f.feedType, '');
      expect(f.amount, '');
      expect(f.notes, '');
      expect(f.fedByName, isNull);
    });
  });

  ({List<Feeding> items, int total}) _result(List<Feeding> items) =>
      (items: items, total: items.length);

  group('load', () {
    final feedings = [_fakeFeeding(id: 1), _fakeFeeding(id: 2)];

    blocTest<FeedingsCubit, FeedingsState>(
      'emits [FeedingsLoading, FeedingsLoaded] on success',
      build: () {
        when(
          () => repo.listFeedings(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => _result(feedings));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<FeedingsLoading>(),
        isA<FeedingsLoaded>()
            .having((s) => s.feedings.length, 'length', 2),
      ],
    );

    blocTest<FeedingsCubit, FeedingsState>(
      'emits [FeedingsLoading, FeedingsError] on failure',
      build: () {
        when(
          () => repo.listFeedings(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenThrow(Exception('network error'));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [isA<FeedingsLoading>(), isA<FeedingsError>()],
    );

    blocTest<FeedingsCubit, FeedingsState>(
      'totalPages is at least 1 when total is 0',
      build: () {
        when(
          () => repo.listFeedings(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => (items: <Feeding>[], total: 0));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<FeedingsLoading>(),
        isA<FeedingsLoaded>()
            .having((s) => s.totalPages, 'totalPages', 1),
      ],
    );
  });

  group('goToPage', () {
    blocTest<FeedingsCubit, FeedingsState>(
      'emits correct currentPage and totalPages',
      build: () {
        when(
          () => repo.listFeedings(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => (items: [_fakeFeeding()], total: 15));
        return cubit;
      },
      act: (c) => c.goToPage(2),
      expect: () => [
        isA<FeedingsLoading>(),
        isA<FeedingsLoaded>()
            .having((s) => s.currentPage, 'currentPage', 2)
            .having((s) => s.totalPages, 'totalPages', 2),
      ],
    );
  });

  group('delete', () {
    blocTest<FeedingsCubit, FeedingsState>(
      'emits [FeedingsLoading, FeedingsLoaded] on success',
      build: () {
        when(
          () => repo.deleteFeeding(apiaryId: 1, hiveId: 10, feedingId: 1),
        ).thenAnswer((_) async {});
        when(
          () => repo.listFeedings(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => _result([]));
        return cubit;
      },
      act: (c) => c.delete(1),
      expect: () => [
        isA<FeedingsLoading>(),
        isA<FeedingsLoaded>()
            .having((s) => s.feedings, 'feedings', isEmpty),
      ],
    );

    blocTest<FeedingsCubit, FeedingsState>(
      'backs up to previous page when current page becomes empty after delete',
      build: () {
        when(
          () => repo.deleteFeeding(apiaryId: 1, hiveId: 10, feedingId: 1),
        ).thenAnswer((_) async {});
        var call = 0;
        when(
          () => repo.listFeedings(1, 10,
              limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async {
          call++;
          if (call == 1) return (items: <Feeding>[], total: 1);
          return (items: [_fakeFeeding()], total: 1);
        });
        return cubit;
      },
      seed: () => FeedingsLoaded([], currentPage: 2, totalPages: 2),
      act: (c) => c.delete(1),
      expect: () => [
        isA<FeedingsLoading>(),
        isA<FeedingsLoading>(),
        isA<FeedingsLoaded>()
            .having((s) => s.currentPage, 'currentPage', 1)
            .having((s) => s.feedings, 'feedings', isNotEmpty),
      ],
    );

    blocTest<FeedingsCubit, FeedingsState>(
      'emits [FeedingsLoading, FeedingsError] on failure',
      build: () {
        when(
          () => repo.deleteFeeding(apiaryId: 1, hiveId: 10, feedingId: 1),
        ).thenThrow(Exception('error'));
        return cubit;
      },
      act: (c) => c.delete(1),
      expect: () => [isA<FeedingsLoading>(), isA<FeedingsError>()],
    );
  });
}
