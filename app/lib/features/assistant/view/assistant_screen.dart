import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_markdown_plus/flutter_markdown_plus.dart';

import '../../../core/api/api_client.dart';
import '../../../core/theme/app_layout.dart';
import '../../../core/widgets/app_drawer.dart';
import '../../../core/widgets/photo_size_snackbar.dart';
import '../../../core/widgets/profile_icon_button.dart';
import '../../../l10n/app_localizations.dart';
import '../cubit/assistant_cubit.dart';
import '../data/assistant_chat_message.dart';
import '../data/assistant_repository.dart';

class AssistantScreen extends StatelessWidget {
  final ValueChanged<AppSection> onSelectSection;

  const AssistantScreen({super.key, required this.onSelectSection});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => AssistantCubit(
        repo: AssistantRepository(api: context.read<ApiClient>()),
      ),
      child: _AssistantView(onSelectSection: onSelectSection),
    );
  }
}

class _AssistantView extends StatefulWidget {
  final ValueChanged<AppSection> onSelectSection;

  const _AssistantView({required this.onSelectSection});

  @override
  State<_AssistantView> createState() => _AssistantViewState();
}

class _AssistantViewState extends State<_AssistantView> {
  final _inputController = TextEditingController();
  final _scrollController = ScrollController();

  @override
  void dispose() {
    _inputController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  void _scrollToBottom() {
    if (!_scrollController.hasClients) return;
    _scrollController.animateTo(
      _scrollController.position.maxScrollExtent,
      duration: const Duration(milliseconds: 200),
      curve: Curves.easeOut,
    );
  }

  void _send(AssistantCubit cubit) {
    final text = _inputController.text.trim();
    if (text.isEmpty) return;
    _inputController.clear();
    cubit.sendMessage(text);
    WidgetsBinding.instance.addPostFrameCallback((_) => _scrollToBottom());
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final cubit = context.read<AssistantCubit>();

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.assistantTitle),
        actions: const [ProfileIconButton()],
      ),
      drawer: AuthenticatedAppDrawer(
        current: AppSection.assistant,
        onSelect: widget.onSelectSection,
      ),
      body: BlocConsumer<AssistantCubit, AssistantState>(
        listener: (context, state) {
          WidgetsBinding.instance.addPostFrameCallback((_) => _scrollToBottom());
          if (state.hasError) {
            showBigSnackBar(context, state.errorMessage ?? l10n.assistantError);
          }
        },
        builder: (context, state) {
          if (state.messages.isEmpty) {
            return Center(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: ConstrainedBox(
                  constraints: BoxConstraints(maxWidth: AppLayout.formConstraints(context).maxWidth),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        l10n.assistantEmpty,
                        textAlign: TextAlign.center,
                        style: Theme.of(context).textTheme.bodyLarge?.copyWith(
                              color: Theme.of(context).colorScheme.onSurfaceVariant,
                            ),
                      ),
                      const SizedBox(height: 24),
                      _MessageInput(
                        controller: _inputController,
                        enabled: !state.isSending,
                        onSend: () => _send(cubit),
                      ),
                    ],
                  ),
                ),
              ),
            );
          }
          return Column(
            children: [
              Expanded(
                child: Center(
                  child: ConstrainedBox(
                    constraints: BoxConstraints(maxWidth: AppLayout.formConstraints(context).maxWidth),
                    child: ListView.builder(
                      controller: _scrollController,
                      padding: const EdgeInsets.all(16),
                      itemCount: state.messages.length,
                      itemBuilder: (context, index) {
                        final message = state.messages[index];
                        final isLast = index == state.messages.length - 1;
                        final isStreamingPlaceholder = isLast &&
                            state.isSending &&
                            message.role == AssistantChatRole.assistant &&
                            message.text.isEmpty;
                        return _ChatBubble(
                          message: message,
                          isPending: isStreamingPlaceholder,
                        );
                      },
                    ),
                  ),
                ),
              ),
              const Divider(height: 1),
              _MessageInput(
                controller: _inputController,
                enabled: !state.isSending,
                onSend: () => _send(cubit),
              ),
            ],
          );
        },
      ),
    );
  }
}

class _ChatBubble extends StatelessWidget {
  final AssistantChatMessage message;
  final bool isPending;

  const _ChatBubble({required this.message, required this.isPending});

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final isUser = message.role == AssistantChatRole.user;
    final bubbleColor = Color.alphaBlend(
      colorScheme.onSurface.withValues(alpha: 0.12),
      Theme.of(context).scaffoldBackgroundColor,
    );

    if (isUser) {
      return Align(
        alignment: Alignment.centerRight,
        child: ConstrainedBox(
          constraints: BoxConstraints(maxWidth: AppLayout.formConstraints(context).maxWidth * 0.7),
          child: Container(
            margin: const EdgeInsets.symmetric(vertical: 6),
            padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
            decoration: BoxDecoration(
              color: bubbleColor,
              borderRadius: BorderRadius.circular(20),
            ),
            child: SelectableText(
              message.text,
              style: TextStyle(color: colorScheme.onSurface),
            ),
          ),
        ),
      );
    }

    return Align(
      alignment: Alignment.centerLeft,
      child: ConstrainedBox(
        constraints: BoxConstraints(maxWidth: AppLayout.formConstraints(context).maxWidth * 0.7),
        child: Container(
          margin: const EdgeInsets.symmetric(vertical: 6),
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
          decoration: BoxDecoration(
            color: bubbleColor,
            borderRadius: BorderRadius.circular(20),
          ),
          child: isPending
              ? SizedBox(
                  width: 28,
                  height: 28,
                  child: CircularProgressIndicator(
                    strokeWidth: 2.5,
                    color: colorScheme.onSurfaceVariant,
                  ),
                )
              : MarkdownBody(
                  data: message.text,
                  selectable: true,
                  styleSheet: MarkdownStyleSheet.fromTheme(Theme.of(context)).copyWith(
                    p: TextStyle(color: colorScheme.onSurface),
                  ),
                ),
        ),
      ),
    );
  }
}

class _MessageInput extends StatelessWidget {
  final TextEditingController controller;
  final bool enabled;
  final VoidCallback onSend;

  const _MessageInput({
    required this.controller,
    required this.enabled,
    required this.onSend,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return SafeArea(
      top: false,
      child: Center(
        child: ConstrainedBox(
          constraints: BoxConstraints(maxWidth: AppLayout.formConstraints(context).maxWidth),
          child: Padding(
            padding: const EdgeInsets.all(8),
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    controller: controller,
                    enabled: enabled,
                    decoration: InputDecoration(
                      hintText: l10n.assistantInputHint,
                      border: const OutlineInputBorder(),
                      isDense: true,
                    ),
                    textInputAction: TextInputAction.send,
                    onSubmitted: (_) => onSend(),
                  ),
                ),
                const SizedBox(width: 8),
                ValueListenableBuilder<TextEditingValue>(
                  valueListenable: controller,
                  builder: (context, value, _) {
                    final canSend = enabled && value.text.trim().isNotEmpty;
                    return IconButton.filled(
                      icon: enabled
                          ? const Icon(Icons.send)
                          : const SizedBox(
                              width: 18,
                              height: 18,
                              child: CircularProgressIndicator(strokeWidth: 2),
                            ),
                      onPressed: canSend ? onSend : null,
                    );
                  },
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
