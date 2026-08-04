import 'package:app/features/voice/data/voice_recording_model.dart';
import 'package:flutter_test/flutter_test.dart';

Map<String, dynamic> _actionJson({
  int id = 1,
  int sequence = 0,
  int? hiveId = 5,
  String? toolName = 'log_inspection',
  String status = 'success',
  String? resultType = 'inspection',
  int? resultRecordId = 9,
  String? errorMessage,
}) =>
    {
      'id': id,
      'sequence': sequence,
      'hive_id': hiveId,
      'tool_name': toolName,
      'status': status,
      'result_type': resultType,
      'result_record_id': resultRecordId,
      'error_message': errorMessage,
    };

Map<String, dynamic> _recordingJson({
  int recordingId = 1,
  String status = 'pending',
  String? transcript = 'checked hive 5',
  String? errorMessage,
  String? createdAt = '2024-05-01T10:00:00Z',
  String? processedAt = '2024-05-01T10:01:00Z',
  List<Map<String, dynamic>>? voiceActions,
}) =>
    {
      'recording_id': recordingId,
      'status': status,
      'transcript': transcript,
      'error_message': errorMessage,
      'created_at': createdAt,
      'processed_at': processedAt,
      'voice_actions': voiceActions ?? [_actionJson()],
    };

void main() {
  group('VoiceAction.fromJson', () {
    test('maps all fields', () {
      final action = VoiceAction.fromJson(_actionJson());

      expect(action.id, 1);
      expect(action.sequence, 0);
      expect(action.hiveId, 5);
      expect(action.toolName, 'log_inspection');
      expect(action.status, 'success');
      expect(action.resultType, 'inspection');
      expect(action.resultRecordId, 9);
      expect(action.errorMessage, isNull);
    });

    test('handles nullable fields', () {
      final action = VoiceAction.fromJson(_actionJson(
        hiveId: null,
        toolName: null,
        resultType: null,
        resultRecordId: null,
        errorMessage: 'boom',
      ));

      expect(action.hiveId, isNull);
      expect(action.toolName, isNull);
      expect(action.resultType, isNull);
      expect(action.resultRecordId, isNull);
      expect(action.errorMessage, 'boom');
    });
  });

  group('VoiceRecording.fromJson', () {
    test('maps all fields including nested voice_actions', () {
      final recording = VoiceRecording.fromJson(_recordingJson());

      expect(recording.recordingId, 1);
      expect(recording.status, 'pending');
      expect(recording.transcript, 'checked hive 5');
      expect(recording.errorMessage, isNull);
      expect(recording.createdAt, isNotNull);
      expect(recording.processedAt, isNotNull);
      expect(recording.voiceActions, hasLength(1));
      expect(recording.voiceActions.single.id, 1);
      expect(recording.localPath, isNull);
    });

    test('handles null transcript/error_message/created_at/processed_at', () {
      final recording = VoiceRecording.fromJson(_recordingJson(
        transcript: null,
        errorMessage: 'failed',
        createdAt: null,
        processedAt: null,
      ));

      expect(recording.transcript, isNull);
      expect(recording.errorMessage, 'failed');
      expect(recording.createdAt, isNull);
      expect(recording.processedAt, isNull);
    });

    test('handles empty voice_actions array', () {
      final recording = VoiceRecording.fromJson(_recordingJson(voiceActions: []));

      expect(recording.voiceActions, isEmpty);
    });
  });

  group('VoiceRecording.copyWith', () {
    test('only overrides status/localPath, leaves everything else unchanged', () {
      final original = VoiceRecording.fromJson(_recordingJson());

      final copy = original.copyWith(status: 'processing', localPath: '/tmp/a.m4a');

      expect(copy.status, 'processing');
      expect(copy.localPath, '/tmp/a.m4a');
      expect(copy.recordingId, original.recordingId);
      expect(copy.transcript, original.transcript);
      expect(copy.errorMessage, original.errorMessage);
      expect(copy.createdAt, original.createdAt);
      expect(copy.processedAt, original.processedAt);
      expect(copy.voiceActions, same(original.voiceActions));
    });

    test('leaves status/localPath unchanged when not provided', () {
      final original = VoiceRecording.fromJson(_recordingJson()).copyWith(
        localPath: '/tmp/a.m4a',
      );

      final copy = original.copyWith();

      expect(copy.status, original.status);
      expect(copy.localPath, original.localPath);
    });

    test('overrides voiceActions when provided', () {
      final original = VoiceRecording.fromJson(_recordingJson());
      final newActions = [
        VoiceAction.fromJson(_actionJson(id: 2, toolName: 'log_treatment')),
      ];

      final copy = original.copyWith(voiceActions: newActions);

      expect(copy.voiceActions, same(newActions));
      expect(copy.voiceActions.single.id, 2);
    });

    test('leaves voiceActions unchanged when not provided', () {
      final original = VoiceRecording.fromJson(_recordingJson());

      final copy = original.copyWith(status: 'processing');

      expect(copy.voiceActions, same(original.voiceActions));
    });
  });
}
