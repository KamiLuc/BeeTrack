import 'package:flutter_test/flutter_test.dart';
import 'package:app/features/honey_batch/data/processing_method.dart';

void main() {
  group('ProcessingMethod', () {
    test('fromJson/toJson round-trip for all values', () {
      for (final method in ProcessingMethod.values) {
        expect(ProcessingMethod.fromJson(method.toJson()), method);
      }
    });

    test('fromJson falls back to raw for unknown value', () {
      expect(ProcessingMethod.fromJson('unknown'), ProcessingMethod.raw);
    });
  });
}
