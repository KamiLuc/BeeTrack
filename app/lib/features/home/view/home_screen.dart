import 'package:flutter/material.dart';

import '../../../core/widgets/profile_icon_button.dart';

class HomeScreen extends StatelessWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('BeeTrack'),
        actions: const [ProfileIconButton()],
      ),
      body: const Center(child: Text('BeeTrack')),
    );
  }
}
