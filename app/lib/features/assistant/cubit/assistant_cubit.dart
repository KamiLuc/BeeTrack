import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../core/api/api_exception.dart';
import '../data/assistant_chat_message.dart';
import '../data/assistant_conversation_summary.dart';
import '../data/assistant_event.dart';
import '../data/assistant_repository.dart';

part 'assistant_state.dart';

const assistantMaxMessagesPerConversation = 10;

class AssistantCubit extends Cubit<AssistantState> {
  final AssistantRepository _repo;

  AssistantCubit({required AssistantRepository repo})
      : _repo = repo,
        super(const AssistantState());

  Future<void> loadConversations() async {
    emit(state.copyWith(isLoadingConversations: true));
    try {
      final conversations = await _repo.fetchConversations();
      emit(state.copyWith(conversations: conversations, isLoadingConversations: false));
    } catch (_) {
      emit(state.copyWith(isLoadingConversations: false));
    }
  }

  void startNewChat() {
    emit(AssistantState(conversations: state.conversations));
  }

  Future<void> openConversation(int conversationId) async {
    if (conversationId == state.conversationId) return;
    emit(state.copyWith(
      messages: const [],
      conversationId: conversationId,
      isLoadingMessages: true,
      hasError: false,
    ));
    try {
      final messages = await _repo.fetchMessages(conversationId);
      emit(state.copyWith(messages: messages, isLoadingMessages: false));
    } on ApiException catch (e) {
      emit(state.copyWith(isLoadingMessages: false, hasError: true, errorMessage: e.message));
    } catch (_) {
      emit(state.copyWith(isLoadingMessages: false, hasError: true));
    }
  }

  Future<void> deleteConversation(int conversationId) async {
    try {
      await _repo.deleteConversation(conversationId);
    } on ApiException catch (e) {
      emit(state.copyWith(hasError: true, errorMessage: e.message));
      return;
    } catch (_) {
      emit(state.copyWith(hasError: true));
      return;
    }

    final remaining = state.conversations.where((c) => c.id != conversationId).toList();
    if (state.conversationId == conversationId) {
      emit(AssistantState(conversations: remaining));
    } else {
      emit(state.copyWith(conversations: remaining));
    }
  }

  Future<void> sendMessage(String text) async {
    if (state.isSending || state.isAtMessageLimit || text.trim().isEmpty) return;

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
