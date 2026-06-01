import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:app/features/apiary/cubit/apiaries_cubit.dart';
import 'package:app/features/apiary/data/apiary_model.dart';
import 'package:app/features/apiary/data/apiary_repository.dart';

class MockApiaryRepository extends Mock implements ApiaryRepository {}

void main() {
  late MockApiaryRepository repo;
  late ApiariesCubit cubit;

  setUp(() {
    repo = MockApiaryRepository();
    cubit = ApiariesCubit(repo: repo);
  });

  tearDown(() => cubit.close());

  group('Apiary.fromJson', () {
    test('parses all fields correctly', () {
      final apiary = Apiary.fromJson({
        'id': 1,
        'name': 'Alpha',
        'lat': 52.0,
        'lng': 21.0,
        'grid_rows': 3,
        'grid_cols': 4,
        'hive_count': 5,
        'user_role': 'owner',
      });
      expect(apiary.id, 1);
      expect(apiary.name, 'Alpha');
      expect(apiary.lat, 52.0);
      expect(apiary.lng, 21.0);
      expect(apiary.gridRows, 3);
      expect(apiary.gridCols, 4);
      expect(apiary.hiveCount, 5);
      expect(apiary.userRole, 'owner');
    });

    test('defaults hive_count to 0 when missing', () {
      final apiary = Apiary.fromJson({
        'id': 1,
        'name': 'Alpha',
        'lat': null,
        'lng': null,
        'grid_rows': 2,
        'grid_cols': 2,
        'user_role': 'member',
      });
      expect(apiary.hiveCount, 0);
      expect(apiary.lat, isNull);
    });
  });

  group('load', () {
    final apiaries = [
      const Apiary(
        id: 1,
        name: 'Alpha',
        lat: 52.0,
        lng: 21.0,
        gridRows: 3,
        gridCols: 4,
        hiveCount: 0,
        userRole: 'owner',
      ),
    ];

    blocTest<ApiariesCubit, ApiariesState>(
      'emits [ApiariesLoading, ApiariesLoaded] on success',
      build: () {
        when(() => repo.listApiaries()).thenAnswer((_) async => apiaries);
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<ApiariesLoading>(),
        isA<ApiariesLoaded>().having((s) => s.apiaries.length, 'length', 1),
      ],
    );

    blocTest<ApiariesCubit, ApiariesState>(
      'emits [ApiariesLoading, ApiariesLoaded] with empty list',
      build: () {
        when(() => repo.listApiaries()).thenAnswer((_) async => []);
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<ApiariesLoading>(),
        isA<ApiariesLoaded>().having((s) => s.apiaries, 'apiaries', isEmpty),
      ],
    );

    blocTest<ApiariesCubit, ApiariesState>(
      'emits [ApiariesLoading, ApiariesError] on failure',
      build: () {
        when(() => repo.listApiaries()).thenThrow(Exception('network error'));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [isA<ApiariesLoading>(), isA<ApiariesError>()],
    );
  });

  group('create', () {
    blocTest<ApiariesCubit, ApiariesState>(
      'emits [ApiariesLoading, ApiariesLoaded] on success',
      build: () {
        when(() => repo.createApiary(
          name: any(named: 'name'),
          lat: any(named: 'lat'),
          lng: any(named: 'lng'),
          gridRows: any(named: 'gridRows'),
          gridCols: any(named: 'gridCols'),
        )).thenAnswer((_) async {});
        when(() => repo.listApiaries()).thenAnswer((_) async => []);
        return cubit;
      },
      act: (c) => c.create(
        name: 'New Apiary',
        gridRows: 3,
        gridCols: 3,
      ),
      expect: () => [isA<ApiariesLoading>(), isA<ApiariesLoaded>()],
    );

    blocTest<ApiariesCubit, ApiariesState>(
      'emits [ApiariesLoading, ApiariesError] when create fails',
      build: () {
        when(() => repo.createApiary(
          name: any(named: 'name'),
          lat: any(named: 'lat'),
          lng: any(named: 'lng'),
          gridRows: any(named: 'gridRows'),
          gridCols: any(named: 'gridCols'),
        )).thenThrow(Exception('error'));
        return cubit;
      },
      act: (c) => c.create(
        name: 'New Apiary',
        gridRows: 3,
        gridCols: 3,
      ),
      expect: () => [isA<ApiariesLoading>(), isA<ApiariesError>()],
    );
  });
}
