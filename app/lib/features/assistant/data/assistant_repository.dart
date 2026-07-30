import 'dart:convert';

import 'package:dio/dio.dart';

import '../../../core/api/api_client.dart';
import '../../../core/api/api_exception.dart';
import 'assistant_chat_message.dart';
import 'assistant_conversation_summary.dart';
import 'assistant_event.dart';

class AssistantRepository {
  final ApiClient _api;

  AssistantRepository({required ApiClient api}) : _api = api;

  Future<List<AssistantConversationSummary>> fetchConversations() async {
    try {
      final response = await _api.dio.get('/api/v1/assistant/conversations');
      final data = response.data['conversations'] as List<dynamic>;
      return data
          .map((e) => AssistantConversationSummary.fromJson(e as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<List<AssistantChatMessage>> fetchMessages(int conversationId) async {
    try {
      final response = await _api.dio.get('/api/v1/assistant/conversations/$conversationId/messages');
      final data = response.data['messages'] as List<dynamic>;
      return data.map((e) {
        final map = e as Map<String, dynamic>;
        final role = map['role'] == 'assistant' ? AssistantChatRole.assistant : AssistantChatRole.user;
        return AssistantChatMessage(role: role, text: map['content'] as String? ?? '');
      }).toList();
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Future<void> deleteConversation(int conversationId) async {
    try {
      await _api.dio.delete('/api/v1/assistant/conversations/$conversationId');
    } on DioException catch (e) {
      throw ApiException.fromDioException(e);
    }
  }

  Stream<AssistantEvent> sendMessage({
    required String message,
    int? conversationId,
  }) async* {
    final Response<ResponseBody> response;
    try {
      response = await _api.dio.post<ResponseBody>(
        '/api/v1/assistant/messages',
        data: {
          if (conversationId != null) 'conversation_id': conversationId,
          'message': message,
        },
        options: Options(responseType: ResponseType.stream),
      );
    } on DioException catch (e) {
      throw await _toApiException(e);
    }

    var buffer = '';
    await for (final chunk in response.data!.stream) {
      buffer += utf8.decode(chunk, allowMalformed: true);
      while (true) {
        final separator = buffer.indexOf('\n\n');
        if (separator == -1) break;
        final raw = buffer.substring(0, separator);
        buffer = buffer.substring(separator + 2);
        final event = _parseEvent(raw);
        if (event != null) yield event;
      }
    }
  }

  AssistantEvent? _parseEvent(String raw) {
    String? eventType;
    final dataLines = <String>[];
    for (final line in raw.split('\n')) {
      if (line.startsWith('event:')) {
        eventType = line.substring('event:'.length).trim();
      } else if (line.startsWith('data:')) {
        dataLines.add(line.substring('data:'.length).trim());
      }
    }
    if (eventType == null || dataLines.isEmpty) return null;

    final data = jsonDecode(dataLines.join('\n')) as Map<String, dynamic>;
    return switch (eventType) {
      'conversation' => AssistantConversationEvent(data['conversation_id'] as int),
      'delta' => AssistantDeltaEvent(data['text'] as String? ?? ''),
      'done' => AssistantDoneEvent(),
      'error' => AssistantErrorEvent(data['message'] as String? ?? ''),
      _ => null,
    };
  }

  // A 4xx/5xx before streaming starts still comes back with responseType
  // stream, so e.response?.data is an undecoded ResponseBody rather than the
  // Map ApiException.fromDioException expects — read and decode it here.
  Future<ApiException> _toApiException(DioException e) async {
    final data = e.response?.data;
    if (data is ResponseBody) {
      try {
        final bytes = await data.stream.fold<List<int>>(
          <int>[],
          (acc, chunk) => acc..addAll(chunk),
        );
        final body = jsonDecode(utf8.decode(bytes)) as Map<String, dynamic>;
        return ApiException(
          code: body['code'] as String? ?? 'UNKNOWN',
          message: body['message'] as String? ?? 'Unknown error',
        );
      } catch (_) {
        return const ApiException(code: 'NETWORK_ERROR', message: 'Network error');
      }
    }
    return ApiException.fromDioException(e);
  }
}
