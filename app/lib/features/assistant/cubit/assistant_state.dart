part of 'assistant_cubit.dart';

class AssistantState {
  final List<AssistantChatMessage> messages;
  final int? conversationId;
  final bool isSending;
  final bool hasError;
  final String? errorMessage;

  const AssistantState({
    this.messages = const [],
    this.conversationId,
    this.isSending = false,
    this.hasError = false,
    this.errorMessage,
  });

  AssistantState copyWith({
    List<AssistantChatMessage>? messages,
    int? conversationId,
    bool? isSending,
    bool? hasError,
    String? errorMessage,
  }) {
    return AssistantState(
      messages: messages ?? this.messages,
      conversationId: conversationId ?? this.conversationId,
      isSending: isSending ?? this.isSending,
      hasError: hasError ?? this.hasError,
      errorMessage: errorMessage ?? this.errorMessage,
    );
  }
}
