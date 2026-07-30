import 'dart:async';

import 'package:bloc_test/bloc_test.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:app/features/assistant/cubit/assistant_cubit.dart';
import 'package:app/features/assistant/data/assistant_chat_message.dart';
import 'package:app/features/assistant/data/assistant_conversation_summary.dart';
import 'package:app/features/assistant/data/assistant_event.dart';
import 'package:app/features/assistant/data/assistant_repository.dart';

class MockAssistantRepository extends Mock implements AssistantRepository {}

void main() {
  late MockAssistantRepository repo;
  late AssistantCubit cubit;

  setUp(() {
    repo = MockAssistantRepository();
    cubit = AssistantCubit(repo: repo);
  });

  tearDown(() => cubit.close());

  void mockStream(List<AssistantEvent> events) {
    when(() => repo.sendMessage(
          message: any(named: 'message'),
          conversationId: any(named: 'conversationId'),
        )).thenAnswer((_) => Stream.fromIterable(events));
  }

  group('sendMessage', () {
    blocTest<AssistantCubit, AssistantState>(
      'appends user and placeholder assistant messages then streams deltas',
      build: () {
        mockStream([
          AssistantDeltaEvent('Hel'),
          AssistantDeltaEvent('lo'),
          AssistantDoneEvent(),
        ]);
        return cubit;
      },
      act: (c) => c.sendMessage('hi'),
      expect: () => [
        isA<AssistantState>()
            .having((s) => s.messages.length, 'messages.length', 2)
            .having((s) => s.messages[0].role, 'user role', AssistantChatRole.user)
            .having((s) => s.messages[0].text, 'user text', 'hi')
            .having((s) => s.messages[1].text, 'assistant text', '')
            .having((s) => s.isSending, 'isSending', true),
        isA<AssistantState>()
            .having((s) => s.messages.last.text, 'assistant text', 'Hel'),
        isA<AssistantState>()
            .having((s) => s.messages.last.text, 'assistant text', 'Hello'),
        isA<AssistantState>()
            .having((s) => s.isSending, 'isSending', false),
      ],
    );

    blocTest<AssistantCubit, AssistantState>(
      'stores conversationId from AssistantConversationEvent for reuse',
      build: () {
        mockStream([
          AssistantConversationEvent(99),
          AssistantDoneEvent(),
        ]);
        return cubit;
      },
      act: (c) => c.sendMessage('hi'),
      verify: (c) {
        expect(c.state.conversationId, 99);
      },
    );

    test('reuses stored conversationId on the next sendMessage call',
        () async {
      mockStream([AssistantConversationEvent(5), AssistantDoneEvent()]);
      await cubit.sendMessage('first');

      when(() => repo.sendMessage(
            message: any(named: 'message'),
            conversationId: any(named: 'conversationId'),
          )).thenAnswer((_) => Stream.fromIterable([AssistantDoneEvent()]));

      await cubit.sendMessage('second');

      verify(() => repo.sendMessage(message: 'second', conversationId: 5))
          .called(1);
    });

    blocTest<AssistantCubit, AssistantState>(
      'sets hasError and stops on AssistantErrorEvent',
      build: () {
        mockStream([
          AssistantDeltaEvent('partial'),
          AssistantErrorEvent('boom'),
          AssistantDeltaEvent('should not apply'),
        ]);
        return cubit;
      },
      act: (c) => c.sendMessage('hi'),
      expect: () => [
        isA<AssistantState>(),
        isA<AssistantState>()
            .having((s) => s.messages.last.text, 'assistant text', 'partial'),
        isA<AssistantState>()
            .having((s) => s.hasError, 'hasError', true)
            .having((s) => s.isSending, 'isSending', false),
      ],
    );

    blocTest<AssistantCubit, AssistantState>(
      'sets hasError when the repo stream throws',
      build: () {
        when(() => repo.sendMessage(
              message: any(named: 'message'),
              conversationId: any(named: 'conversationId'),
            )).thenAnswer((_) => Stream.error(Exception('network error')));
        return cubit;
      },
      act: (c) => c.sendMessage('hi'),
      expect: () => [
        isA<AssistantState>(),
        isA<AssistantState>()
            .having((s) => s.hasError, 'hasError', true)
            .having((s) => s.isSending, 'isSending', false),
      ],
    );

    blocTest<AssistantCubit, AssistantState>(
      'is a no-op for empty/whitespace-only input',
      build: () => cubit,
      act: (c) => c.sendMessage('   '),
      expect: () => [],
      verify: (_) {
        verifyNever(() => repo.sendMessage(
              message: any(named: 'message'),
              conversationId: any(named: 'conversationId'),
            ));
      },
    );

    test('is a no-op for concurrent sends while already sending', () async {
      final controller = StreamController<AssistantEvent>();
      when(() => repo.sendMessage(
            message: any(named: 'message'),
            conversationId: any(named: 'conversationId'),
          )).thenAnswer((_) => controller.stream);

      final first = cubit.sendMessage('one');
      await Future.delayed(Duration.zero);
      await cubit.sendMessage('two');

      verify(() => repo.sendMessage(
            message: 'one',
            conversationId: any(named: 'conversationId'),
          )).called(1);
      verifyNever(() => repo.sendMessage(
            message: 'two',
            conversationId: any(named: 'conversationId'),
          ));

      controller.add(AssistantDoneEvent());
      await controller.close();
      await first;
    });

    blocTest<AssistantCubit, AssistantState>(
      'is a no-op once the message limit is reached',
      build: () => cubit,
      seed: () => AssistantState(
        messages: List.generate(
          assistantMaxMessagesPerConversation,
          (_) => const AssistantChatMessage(role: AssistantChatRole.user, text: 'x'),
        ),
      ),
      act: (c) => c.sendMessage('one more'),
      expect: () => [],
      verify: (_) {
        verifyNever(() => repo.sendMessage(
              message: any(named: 'message'),
              conversationId: any(named: 'conversationId'),
            ));
      },
    );
  });

  group('loadConversations', () {
    blocTest<AssistantCubit, AssistantState>(
      'populates conversations on success',
      build: () {
        when(() => repo.fetchConversations()).thenAnswer((_) async => [
              AssistantConversationSummary(
                id: 1,
                createdAt: DateTime(2026, 1, 1),
                preview: 'hi',
                messageCount: 2,
              ),
            ]);
        return cubit;
      },
      act: (c) => c.loadConversations(),
      expect: () => [
        isA<AssistantState>().having((s) => s.isLoadingConversations, 'isLoadingConversations', true),
        isA<AssistantState>()
            .having((s) => s.isLoadingConversations, 'isLoadingConversations', false)
            .having((s) => s.conversations.length, 'conversations.length', 1),
      ],
    );

    blocTest<AssistantCubit, AssistantState>(
      'clears the loading flag when the repo call fails',
      build: () {
        when(() => repo.fetchConversations()).thenThrow(Exception('boom'));
        return cubit;
      },
      act: (c) => c.loadConversations(),
      expect: () => [
        isA<AssistantState>().having((s) => s.isLoadingConversations, 'isLoadingConversations', true),
        isA<AssistantState>().having((s) => s.isLoadingConversations, 'isLoadingConversations', false),
      ],
    );
  });

  group('startNewChat', () {
    blocTest<AssistantCubit, AssistantState>(
      'resets messages and conversationId but keeps the conversation list',
      build: () => cubit,
      seed: () => const AssistantState(
        messages: [AssistantChatMessage(role: AssistantChatRole.user, text: 'hi')],
        conversationId: 5,
      ),
      act: (c) => c.startNewChat(),
      expect: () => [
        isA<AssistantState>()
            .having((s) => s.messages, 'messages', isEmpty)
            .having((s) => s.conversationId, 'conversationId', isNull),
      ],
    );
  });

  group('openConversation', () {
    blocTest<AssistantCubit, AssistantState>(
      'loads and shows the conversation history',
      build: () {
        when(() => repo.fetchMessages(7)).thenAnswer((_) async => [
              const AssistantChatMessage(role: AssistantChatRole.user, text: 'hi'),
              const AssistantChatMessage(role: AssistantChatRole.assistant, text: 'hello'),
            ]);
        return cubit;
      },
      act: (c) => c.openConversation(7),
      expect: () => [
        isA<AssistantState>()
            .having((s) => s.isLoadingMessages, 'isLoadingMessages', true)
            .having((s) => s.conversationId, 'conversationId', 7),
        isA<AssistantState>()
            .having((s) => s.isLoadingMessages, 'isLoadingMessages', false)
            .having((s) => s.messages.length, 'messages.length', 2),
      ],
    );

    blocTest<AssistantCubit, AssistantState>(
      'is a no-op when the conversation is already open',
      build: () => cubit,
      seed: () => const AssistantState(conversationId: 7),
      act: (c) => c.openConversation(7),
      expect: () => [],
      verify: (_) {
        verifyNever(() => repo.fetchMessages(any()));
      },
    );

    blocTest<AssistantCubit, AssistantState>(
      'sets hasError when loading history fails',
      build: () {
        when(() => repo.fetchMessages(7)).thenThrow(Exception('boom'));
        return cubit;
      },
      act: (c) => c.openConversation(7),
      expect: () => [
        isA<AssistantState>().having((s) => s.isLoadingMessages, 'isLoadingMessages', true),
        isA<AssistantState>()
            .having((s) => s.isLoadingMessages, 'isLoadingMessages', false)
            .having((s) => s.hasError, 'hasError', true),
      ],
    );
  });

  group('deleteConversation', () {
    blocTest<AssistantCubit, AssistantState>(
      'removes the conversation from the list',
      build: () {
        when(() => repo.deleteConversation(1)).thenAnswer((_) async {});
        return cubit;
      },
      seed: () => AssistantState(
        conversations: [
          AssistantConversationSummary(id: 1, createdAt: DateTime(2026, 1, 1), preview: 'a', messageCount: 2),
          AssistantConversationSummary(id: 2, createdAt: DateTime(2026, 1, 2), preview: 'b', messageCount: 2),
        ],
      ),
      act: (c) => c.deleteConversation(1),
      expect: () => [
        isA<AssistantState>()
            .having((s) => s.conversations.map((e) => e.id).toList(), 'conversation ids', [2]),
      ],
    );

    blocTest<AssistantCubit, AssistantState>(
      'resets the open chat when deleting the currently open conversation',
      build: () {
        when(() => repo.deleteConversation(1)).thenAnswer((_) async {});
        return cubit;
      },
      seed: () => AssistantState(
        conversationId: 1,
        messages: const [AssistantChatMessage(role: AssistantChatRole.user, text: 'hi')],
        conversations: [
          AssistantConversationSummary(id: 1, createdAt: DateTime(2026, 1, 1), preview: 'a', messageCount: 2),
        ],
      ),
      act: (c) => c.deleteConversation(1),
      expect: () => [
        isA<AssistantState>()
            .having((s) => s.conversationId, 'conversationId', isNull)
            .having((s) => s.messages, 'messages', isEmpty)
            .having((s) => s.conversations, 'conversations', isEmpty),
      ],
    );

    blocTest<AssistantCubit, AssistantState>(
      'sets hasError when the repo call fails',
      build: () {
        when(() => repo.deleteConversation(1)).thenThrow(Exception('boom'));
        return cubit;
      },
      act: (c) => c.deleteConversation(1),
      expect: () => [
        isA<AssistantState>().having((s) => s.hasError, 'hasError', true),
      ],
    );
  });
}
