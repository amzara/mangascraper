import 'manga.dart';
import 'chapter.dart';

class MangaDetail {
  final Manga manga;
  final List<Chapter> chapters;

  MangaDetail({
    required this.manga,
    required this.chapters,
  });

  factory MangaDetail.fromJson(Map<String, dynamic> json) {
    final mangaJson = json['manga'] as Map<String, dynamic>;
    final chaptersJson = json['chapters'] as List<dynamic>;

    return MangaDetail(
      manga: Manga.fromJson(mangaJson),
      chapters: chaptersJson.map((c) => Chapter.fromJson(c)).toList(),
    );
  }

  // Get downloaded count
  int get downloadedCount =>
      chapters.where((c) => c.status.toLowerCase() == 'downloaded').length;

  // Get total count
  int get totalCount => chapters.length;
}
