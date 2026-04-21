class SearchResult {
  final int id;
  final String author;
  final String name;
  final String chapterLatest;
  final String url;
  final String thumb;
  final String slug;

  SearchResult({
    required this.id,
    required this.author,
    required this.name,
    required this.chapterLatest,
    required this.url,
    required this.thumb,
    required this.slug,
  });

  factory SearchResult.fromJson(Map<String, dynamic> json) {
    return SearchResult(
      id: json['id'] as int,
      author: json['author'] as String,
      name: json['name'] as String,
      chapterLatest: json['chapterLatest'] as String,
      url: json['url'] as String,
      thumb: json['thumb'] as String,
      slug: json['slug'] as String,
    );
  }
}
