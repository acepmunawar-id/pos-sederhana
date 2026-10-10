import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../../../core/providers.dart';

class StoreSettings {
  const StoreSettings({required this.name, required this.taxRateBP, this.address, this.phone, this.logoUrl, this.receiptFooter});

  final String name;
  final String? address;
  final String? phone;
  final String? logoUrl;
  final int taxRateBP;
  final String? receiptFooter;

  factory StoreSettings.fromJson(Map<String, dynamic> j) => StoreSettings(
        name: j['name'] as String,
        address: j['address'] as String?,
        phone: j['phone'] as String?,
        logoUrl: j['logo_url'] as String?,
        taxRateBP: j['tax_rate_bp'] as int,
        receiptFooter: j['receipt_footer'] as String?,
      );
}

class SettingsRepository {
  SettingsRepository(this._api);
  final ApiClient _api;

  Future<StoreSettings> get() => _api.get('/settings', StoreSettings.fromJson);
}

final settingsRepositoryProvider = Provider((ref) => SettingsRepository(ref.watch(apiClientProvider)));
