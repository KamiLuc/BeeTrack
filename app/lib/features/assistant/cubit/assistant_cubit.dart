import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_exception.dart';
import '../data/assistant_chat_message.dart';
import '../data/assistant_event.dart';
import '../data/assistant_repository.dart';

part 'assistant_state.dart';

class AssistantCubit extends Cubit<AssistantState> {
  final AssistantRepository _repo;

  AssistantCubit({required AssistantRepository repo})
      : _repo = repo,
        super(const AssistantState());

  Future<void> sendMessage(String text) async {
    if (state.isSending || text.trim().isEmpty) return;

    emit(state.copyWith(
      messages: [
        ...state.messages,
        AssistantChatMessage(role: AssistantChatRole.user, text: text),
        const AssistantChatMessage(role: AssistantChatRole.assistant, text: ''),
      ],
      isSending: true,
      hasError: false,
    ));

    try {
      await for (final event in _repo.sendMessage(
        message: text,
        conversationId: state.conversationId,
      )) {
        switch (event) {
          case AssistantConversationEvent(:final conversationId):
            emit(state.copyWith(conversationId: conversationId));
          case AssistantDeltaEvent(:final text):
            _appendToLastMessage(text);
          case AssistantDoneEvent():
            break;
          case AssistantErrorEvent(:final message):
            emit(state.copyWith(isSending: false, hasError: true, errorMessage: message));
            return;
        }
      }
      emit(state.copyWith(isSending: false));
    } on ApiException catch (e) {
      emit(state.copyWith(isSending: false, hasError: true, errorMessage: e.message));
    } catch (_) {
      emit(state.copyWith(isSending: false, hasError: true));
    }
  }

  void _appendToLastMessage(String delta) {
    final messages = [...state.messages];
    final last = messages.removeLast();
    messages.add(last.copyWith(text: last.text + delta));
    emit(state.copyWith(messages: messages));
  }
}
