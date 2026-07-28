enum AssistantChatRole { user, assistant }

class AssistantChatMessage {
  final AssistantChatRole role;
  final String text;

  const AssistantChatMessage({required this.role, required this.text});

  AssistantChatMessage copyWith({String? text}) =>
      AssistantChatMessage(role: role, text: text ?? this.text);
}
