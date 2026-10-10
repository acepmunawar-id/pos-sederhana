import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/settings_repository.dart';

final storeSettingsProvider = FutureProvider<StoreSettings>((ref) => ref.watch(settingsRepositoryProvider).get());
