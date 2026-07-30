class AssistantConversationSummary {
  final int id;
  final DateTime createdAt;
  final String preview;
  final int messageCount;

  const AssistantConversationSummary({
    required this.id,
    required this.createdAt,
    required this.preview,
    required this.messageCount,
  });

  factory AssistantConversationSummary.fromJson(Map<String, dynamic> json) {
    return AssistantConversationSummary(
      id: json['id'] as int,
      createdAt: DateTime.parse(json['created_at'] as String),
      preview: json['preview'] as String? ?? '',
      messageCount: json['message_count'] as int? ?? 0,
    );
  }
}
