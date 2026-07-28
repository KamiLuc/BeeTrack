sealed class AssistantEvent {}

class AssistantConversationEvent extends AssistantEvent {
  final int conversationId;

  AssistantConversationEvent(this.conversationId);
}

class AssistantDeltaEvent extends AssistantEvent {
  final String text;

  AssistantDeltaEvent(this.text);
}

class AssistantDoneEvent extends AssistantEvent {}

class AssistantErrorEvent extends AssistantEvent {
  final String message;

  AssistantErrorEvent(this.message);
}
