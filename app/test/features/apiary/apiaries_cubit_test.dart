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

  group('load', () {
    final apiaries = [
      const Apiary(
        id: 1,
        name: 'Alpha',
        lat: 52.0,
        lng: 21.0,
        gridRows: 3,
        gridCols: 4,
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
}
