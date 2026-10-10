class AppUser {
  const AppUser({required this.id, required this.name, required this.email, required this.roleCode, required this.roleName});

  final int id;
  final String name;
  final String email;
  final String roleCode; // owner | admin | cashier
  final String roleName;

  bool get isManager => roleCode == 'owner' || roleCode == 'admin';

  String get initials {
    final parts = name.trim().split(RegExp(r'\s+')).where((p) => p.isNotEmpty).take(2);
    return parts.map((p) => p[0].toUpperCase()).join();
  }

  factory AppUser.fromJson(Map<String, dynamic> j) {
    final role = j['role'] as Map<String, dynamic>;
    return AppUser(
      id: j['id'] as int,
      name: j['name'] as String,
      email: j['email'] as String,
      roleCode: role['code'] as String,
      roleName: role['name'] as String,
    );
  }
}
