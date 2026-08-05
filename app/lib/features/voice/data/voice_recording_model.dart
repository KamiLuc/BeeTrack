const voiceRecordingStatusPending = 'pending';
const voiceRecordingStatusProcessing = 'processing';
const voiceRecordingStatusCompleted = 'completed';
const voiceRecordingStatusFailed = 'failed';

class VoiceAction {
  final int id;
  final int sequence;
  final int? hiveId;
  final String? toolName;
  final Map<String, dynamic>? toolArguments;
  final String status;
  final String? resultType;
  final int? resultRecordId;
  final String? errorMessage;

  const VoiceAction({
    required this.id,
    required this.sequence,
    this.hiveId,
    this.toolName,
    this.toolArguments,
    required this.status,
    this.resultType,
    this.resultRecordId,
    this.errorMessage,
  });

  factory VoiceAction.fromJson(Map<String, dynamic> json) {
    return VoiceAction(
      id: json['id'] as int,
      sequence: json['sequence'] as int,
      hiveId: json['hive_id'] as int?,
      toolName: json['tool_name'] as String?,
      toolArguments: json['tool_arguments'] as Map<String, dynamic>?,
      status: json['status'] as String,
      resultType: json['result_type'] as String?,
      resultRecordId: json['result_record_id'] as int?,
      errorMessage: json['error_message'] as String?,
    );
  }
}

class VoiceRecording {
  final int recordingId;
  final String status;
  final String? transcript;
  final String? errorMessage;
  final DateTime? createdAt;
  final DateTime? processedAt;
  final List<VoiceAction> voiceActions;

  /// Path (or, on web, a blob object URL) to the recording's audio file on
  /// this device. Only ever set client-side for a recording captured in the
  /// current dialog session — the backend never serves audio back, so a
  /// recording fetched fresh from the server has no local audio to play.
  final String? localPath;

  const VoiceRecording({
    required this.recordingId,
    required this.status,
    this.transcript,
    this.errorMessage,
    this.createdAt,
    this.processedAt,
    this.voiceActions = const [],
    this.localPath,
  });

  factory VoiceRecording.fromJson(Map<String, dynamic> json) {
    return VoiceRecording(
      recordingId: json['recording_id'] as int,
      status: json['status'] as String,
      transcript: json['transcript'] as String?,
      errorMessage: json['error_message'] as String?,
      createdAt: json['created_at'] != null
          ? DateTime.parse(json['created_at'] as String).toLocal()
          : null,
      processedAt: json['processed_at'] != null
          ? DateTime.parse(json['processed_at'] as String).toLocal()
          : null,
      voiceActions: (json['voice_actions'] as List<dynamic>? ?? [])
          .map((e) => VoiceAction.fromJson(e as Map<String, dynamic>))
          .toList(),
    );
  }

  VoiceRecording copyWith({
    String? status,
    String? localPath,
    List<VoiceAction>? voiceActions,
  }) {
    return VoiceRecording(
      recordingId: recordingId,
      status: status ?? this.status,
      transcript: transcript,
      errorMessage: errorMessage,
      createdAt: createdAt,
      processedAt: processedAt,
      voiceActions: voiceActions ?? this.voiceActions,
      localPath: localPath ?? this.localPath,
    );
  }
}
