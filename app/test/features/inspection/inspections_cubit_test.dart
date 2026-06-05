import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:app/features/inspection/cubit/inspections_cubit.dart';
import 'package:app/features/inspection/data/inspection_image_model.dart';
import 'package:app/features/inspection/data/inspection_model.dart';
import 'package:app/features/inspection/data/inspection_repository.dart';

class MockInspectionRepository extends Mock implements InspectionRepository {}

Inspection _fakeInspection({int id = 1}) => Inspection(
  id: id,
  hiveId: 10,
  inspectedAt: DateTime(2025, 6, 1),
  queenSeen: 'seen',
  broodPattern: 'good',
  aggressiveness: 'calm',
  queenAdded: false,
  notes: '',
);

void main() {
  late MockInspectionRepository repo;
  late InspectionsCubit cubit;

  setUp(() {
    repo = MockInspectionRepository();
    cubit = InspectionsCubit(repo: repo, apiaryId: 1, hiveId: 10);
  });

  tearDown(() => cubit.close());

  group('InspectionImage.fromJson', () {
    test('parses all fields correctly', () {
      final json = {
        'id': 7,
        'inspection_id': 42,
        'mime_type': 'image/jpeg',
        'created_at': '2025-06-01T10:00:00Z',
      };
      final img = InspectionImage.fromJson(json);
      expect(img.id, 7);
      expect(img.inspectionId, 42);
      expect(img.mimeType, 'image/jpeg');
      expect(img.createdAt, DateTime.utc(2025, 6, 1, 10));
    });
  });

  group('Inspection.fromJson', () {
    test('parses all fields correctly', () {
      final json = {
        'id': 42,
        'hive_id': 10,
        'inspected_at': '2025-06-01T10:00:00Z',
        'queen_status': 'seen',
        'brood_pattern': 'good',
        'aggressiveness': 'calm',
        'frames_brood': 6,
        'frames_honey': 5,
        'frames_pollen': 2,
        'frames_added_drawn': 1,
        'frames_added_foundation': 0,
        'frames_added_honey': 3,
        'queen_cells_count': 0,
        'queen_added': false,
        'notes': 'All good',
        'diseases': [],
      };
      final insp = Inspection.fromJson(json);
      expect(insp.id, 42);
      expect(insp.hiveId, 10);
      expect(insp.queenSeen, 'seen');
      expect(insp.broodPattern, 'good');
      expect(insp.aggressiveness, 'calm');
      expect(insp.framesBrood, 6);
      expect(insp.framesHoney, 5);
      expect(insp.framesAddedHoney, 3);
      expect(insp.queenAdded, false);
      expect(insp.notes, 'All good');
    });

    test('handles null optional fields', () {
      final json = {
        'id': 1,
        'hive_id': 10,
        'inspected_at': '2025-06-01T10:00:00Z',
        'queen_status': null,
        'brood_pattern': null,
        'aggressiveness': null,
        'frames_brood': null,
        'frames_honey': null,
        'frames_added_honey': null,
        'queen_added': false,
        'notes': null,
      };
      final insp = Inspection.fromJson(json);
      expect(insp.queenSeen, '');
      expect(insp.framesBrood, isNull);
      expect(insp.framesHoney, isNull);
      expect(insp.framesAddedHoney, isNull);
      expect(insp.notes, '');
    });
  });

  group('load', () {
    final inspections = [_fakeInspection(id: 1), _fakeInspection(id: 2)];

    blocTest<InspectionsCubit, InspectionsState>(
      'emits [InspectionsLoading, InspectionsLoaded] on success',
      build: () {
        when(
          () => repo.listInspections(1, 10),
        ).thenAnswer((_) async => inspections);
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<InspectionsLoading>(),
        isA<InspectionsLoaded>().having(
          (s) => s.inspections.length,
          'length',
          2,
        ),
      ],
    );

    blocTest<InspectionsCubit, InspectionsState>(
      'emits [InspectionsLoading, InspectionsLoaded] with empty list',
      build: () {
        when(() => repo.listInspections(1, 10)).thenAnswer((_) async => []);
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<InspectionsLoading>(),
        isA<InspectionsLoaded>().having(
          (s) => s.inspections,
          'inspections',
          isEmpty,
        ),
      ],
    );

    blocTest<InspectionsCubit, InspectionsState>(
      'emits [InspectionsLoading, InspectionsError] on failure',
      build: () {
        when(
          () => repo.listInspections(1, 10),
        ).thenThrow(Exception('network error'));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [isA<InspectionsLoading>(), isA<InspectionsError>()],
    );
  });

  group('delete', () {
    blocTest<InspectionsCubit, InspectionsState>(
      'emits [InspectionsLoading, InspectionsLoaded] on success',
      build: () {
        when(
          () => repo.deleteInspection(apiaryId: 1, hiveId: 10, inspectionId: 1),
        ).thenAnswer((_) async {});
        when(() => repo.listInspections(1, 10)).thenAnswer((_) async => []);
        return cubit;
      },
      act: (c) => c.delete(1),
      expect: () => [
        isA<InspectionsLoading>(),
        isA<InspectionsLoaded>().having(
          (s) => s.inspections,
          'inspections',
          isEmpty,
        ),
      ],
    );

    blocTest<InspectionsCubit, InspectionsState>(
      'emits [InspectionsLoading, InspectionsError] on failure',
      build: () {
        when(
          () => repo.deleteInspection(apiaryId: 1, hiveId: 10, inspectionId: 1),
        ).thenThrow(Exception('error'));
        return cubit;
      },
      act: (c) => c.delete(1),
      expect: () => [isA<InspectionsLoading>(), isA<InspectionsError>()],
    );
  });
}
