class Manga {
  final int mangaId;
  final String title;
  final String slug;
  final DateTime createdAt;

  Manga({
    required this.mangaId,
    required this.title,
    required this.slug,
    required this.createdAt,
  });

  factory Manga.fromJson(Map<String, dynamic> json) {
    return Manga(
      mangaId: json['MangaID'] as int,
      title: json['Title'] as String,
      slug: json['Slug'] as String,
      createdAt: DateTime.parse(json['CreatedAt'] as String),
    );
  }
}
