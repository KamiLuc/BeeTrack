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
        'frames_feed': 5,
        'frames_pollen': 2,
        'frames_added_drawn': -1,
        'frames_added_foundation': 0,
        'frames_added_brood': 4,
        'frames_added_feed': -3,
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
      expect(insp.framesFeed, 5);
      expect(insp.framesAddedDrawn, -1);
      expect(insp.framesAddedFoundation, 0);
      expect(insp.framesAddedBrood, 4);
      expect(insp.framesAddedFeed, -3);
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
        'frames_feed': null,
        'frames_added_feed': null,
        'queen_added': false,
        'notes': null,
      };
      final insp = Inspection.fromJson(json);
      expect(insp.queenSeen, '');
      expect(insp.framesBrood, isNull);
      expect(insp.framesFeed, isNull);
      expect(insp.framesAddedFeed, isNull);
      expect(insp.notes, '');
    });

    test('parses inspected_by_name when present', () {
      final json = {
        'id': 1,
        'hive_id': 10,
        'inspected_at': '2025-06-01T10:00:00Z',
        'queen_added': false,
        'notes': '',
        'inspected_by_name': 'Alice',
      };
      final insp = Inspection.fromJson(json);
      expect(insp.inspectedByName, 'Alice');
    });

    test('sets inspectedByName to null when absent from json', () {
      final json = {
        'id': 1,
        'hive_id': 10,
        'inspected_at': '2025-06-01T10:00:00Z',
        'queen_added': false,
        'notes': '',
      };
      final insp = Inspection.fromJson(json);
      expect(insp.inspectedByName, isNull);
    });
  });

  ({List<Inspection> items, int total}) _result(List<Inspection> items) =>
      (items: items, total: items.length);

  group('load', () {
    final inspections = [_fakeInspection(id: 1), _fakeInspection(id: 2)];

    blocTest<InspectionsCubit, InspectionsState>(
      'emits [InspectionsLoading, InspectionsLoaded] on success',
      build: () {
        when(
          () => repo.listInspections(1, 10, limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => _result(inspections));
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
        when(
          () => repo.listInspections(1, 10, limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => _result([]));
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
          () => repo.listInspections(1, 10, limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenThrow(Exception('network error'));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [isA<InspectionsLoading>(), isA<InspectionsError>()],
    );
  });

  group('goToPage', () {
    blocTest<InspectionsCubit, InspectionsState>(
      'emits correct currentPage and totalPages',
      build: () {
        when(
          () => repo.listInspections(1, 10, limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => (items: [_fakeInspection()], total: 15));
        return cubit;
      },
      act: (c) => c.goToPage(2),
      expect: () => [
        isA<InspectionsLoading>(),
        isA<InspectionsLoaded>()
            .having((s) => s.currentPage, 'currentPage', 2)
            .having((s) => s.totalPages, 'totalPages', 2),
      ],
    );

    blocTest<InspectionsCubit, InspectionsState>(
      'totalPages is at least 1 when total is 0',
      build: () {
        when(
          () => repo.listInspections(1, 10, limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => (items: <Inspection>[], total: 0));
        return cubit;
      },
      act: (c) => c.load(),
      expect: () => [
        isA<InspectionsLoading>(),
        isA<InspectionsLoaded>().having((s) => s.totalPages, 'totalPages', 1),
      ],
    );
  });

  group('delete', () {
    blocTest<InspectionsCubit, InspectionsState>(
      'emits [InspectionsLoading, InspectionsLoaded] on success',
      build: () {
        when(
          () => repo.deleteInspection(apiaryId: 1, hiveId: 10, inspectionId: 1),
        ).thenAnswer((_) async {});
        when(
          () => repo.listInspections(1, 10, limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async => _result([]));
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
      'backs up to previous page when current page becomes empty after delete',
      build: () {
        when(
          () => repo.deleteInspection(apiaryId: 1, hiveId: 10, inspectionId: 1),
        ).thenAnswer((_) async {});
        var call = 0;
        when(
          () => repo.listInspections(1, 10, limit: any(named: 'limit'), offset: any(named: 'offset')),
        ).thenAnswer((_) async {
          call++;
          if (call == 1) return (items: <Inspection>[], total: 1);
          return (items: [_fakeInspection()], total: 1);
        });
        return cubit;
      },
      seed: () => InspectionsLoaded([], currentPage: 2, totalPages: 2),
      act: (c) => c.delete(1),
      expect: () => [
        isA<InspectionsLoading>(),
        isA<InspectionsLoading>(),
        isA<InspectionsLoaded>()
            .having((s) => s.currentPage, 'currentPage', 1)
            .having((s) => s.inspections, 'inspections', isNotEmpty),
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
