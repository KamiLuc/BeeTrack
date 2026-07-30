part of 'assistant_cubit.dart';

class AssistantState {
  final List<AssistantChatMessage> messages;
  final int? conversationId;
  final bool isSending;
  final bool hasError;
  final String? errorMessage;
  final List<AssistantConversationSummary> conversations;
  final bool isLoadingConversations;
  final bool isLoadingMessages;

  const AssistantState({
    this.messages = const [],
    this.conversationId,
    this.isSending = false,
    this.hasError = false,
    this.errorMessage,
    this.conversations = const [],
    this.isLoadingConversations = false,
    this.isLoadingMessages = false,
  });

  bool get isAtMessageLimit => messages.length >= assistantMaxMessagesPerConversation;

  AssistantState copyWith({
    List<AssistantChatMessage>? messages,
    int? conversationId,
    bool? isSending,
    bool? hasError,
    String? errorMessage,
    List<AssistantConversationSummary>? conversations,
    bool? isLoadingConversations,
    bool? isLoadingMessages,
  }) {
    return AssistantState(
      messages: messages ?? this.messages,
      conversationId: conversationId ?? this.conversationId,
      isSending: isSending ?? this.isSending,
      hasError: hasError ?? this.hasError,
      errorMessage: errorMessage ?? this.errorMessage,
      conversations: conversations ?? this.conversations,
      isLoadingConversations: isLoadingConversations ?? this.isLoadingConversations,
      isLoadingMessages: isLoadingMessages ?? this.isLoadingMessages,
    );
  }
}
