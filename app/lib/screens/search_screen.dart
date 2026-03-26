import 'dart:async';
import 'package:flutter/material.dart';
import '../models/search_result.dart';
import '../services/api_service.dart';

class SearchScreen extends StatefulWidget {
  const SearchScreen({super.key});

  @override
  State<SearchScreen> createState() => _SearchScreenState();
}

class _SearchScreenState extends State<SearchScreen> {
  final ApiService _apiService = ApiService();
  final TextEditingController _searchController = TextEditingController();
  final FocusNode _searchFocusNode = FocusNode();
  Timer? _debounceTimer;
  List<SearchResult> _searchResults = [];
  bool _isLoading = false;
  String? _error;
  
  // Token state
  bool _isGettingToken = true;
  bool _hasToken = false;
  String? _tokenError;

  @override
  void initState() {
    super.initState();
    _getToken();
  }

  @override
  void dispose() {
    _debounceTimer?.cancel();
    _searchController.dispose();
    _searchFocusNode.dispose();
    super.dispose();
  }

  Future<void> _getToken() async {
    setState(() {
      _isGettingToken = true;
      _tokenError = null;
    });

    try {
      await _apiService.getToken();
      setState(() {
        _isGettingToken = false;
        _hasToken = true;
      });
      // Auto-focus search field after getting token
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _searchFocusNode.requestFocus();
      });
    } catch (e) {
      setState(() {
        _isGettingToken = false;
        _hasToken = false;
        _tokenError = e.toString();
      });
    }
  }

  void _onSearchChanged(String query) {
    _debounceTimer?.cancel();

    if (query.isEmpty) {
      setState(() {
        _searchResults = [];
        _error = null;
      });
      return;
    }

    _debounceTimer = Timer(const Duration(seconds: 1), () {
      _performSearch(query);
    });
  }

  Future<void> _performSearch(String query) async {
    if (query.isEmpty) return;

    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final results = await _apiService.searchManga(query);
      setState(() {
        _searchResults = results;
        _isLoading = false;
      });
    } catch (e) {
      final errorMsg = e.toString();
      
      // Check if error is about missing token
      if (errorMsg.contains('token not available') || errorMsg.contains('POST /api/token')) {
        // Token expired or not available, retry getting it
        setState(() {
          _hasToken = false;
          _tokenError = 'Token expired. Please retry.';
          _isLoading = false;
        });
        _getToken();
        return;
      }
      
      setState(() {
        _error = errorMsg;
        _isLoading = false;
      });
    }
  }

  Future<void> _showDownloadConfirmDialog(SearchResult manga) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Download Manga?'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              manga.name,
              style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
            ),
            const SizedBox(height: 8),
            Text('Author: ${manga.author}'),
            Text('Latest: ${manga.chapterLatest}'),
            const SizedBox(height: 16),
            const Text('Do you want to download this manga?'),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () => Navigator.pop(context, true),
            child: const Text('Download'),
          ),
        ],
      ),
    );

    if (confirmed == true) {
      await _downloadManga(manga);
    }
  }

  Future<void> _downloadManga(SearchResult manga) async {
    try {
      await _apiService.downloadManga(manga.slug);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Started downloading: ${manga.name}'),
            backgroundColor: Colors.green,
          ),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Add Manga'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
      ),
      body: _buildBody(),
    );
  }

  Widget _buildBody() {
    // Step 1: Getting token
    if (_isGettingToken) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            CircularProgressIndicator(),
            SizedBox(height: 16),
            Text('Getting access token...'),
            SizedBox(height: 8),
            Text(
              'Please wait while we connect to the manga source',
              style: TextStyle(color: Colors.grey, fontSize: 12),
            ),
          ],
        ),
      );
    }

    // Step 2: Token error
    if (!_hasToken) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.error_outline, color: Colors.red, size: 48),
            const SizedBox(height: 16),
            const Text(
              'Failed to get access token',
              style: TextStyle(fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            Text(
              _tokenError ?? 'Unknown error',
              textAlign: TextAlign.center,
              style: const TextStyle(color: Colors.grey),
            ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: _getToken,
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
          ],
        ),
      );
    }

    // Step 3: Token ready - show search UI
    return Column(
      children: [
        // Success banner
        Container(
          width: double.infinity,
          color: Colors.green.shade50,
          padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 16),
          child: Row(
            children: [
              Icon(Icons.check_circle, color: Colors.green.shade700, size: 16),
              const SizedBox(width: 8),
              Text(
                'Access token retrieved',
                style: TextStyle(color: Colors.green.shade700, fontSize: 12),
              ),
            ],
          ),
        ),

        // Search box
        Padding(
          padding: const EdgeInsets.all(16),
          child: TextField(
            controller: _searchController,
            focusNode: _searchFocusNode,
            onChanged: _onSearchChanged,
            decoration: InputDecoration(
              hintText: 'Search for manga...',
              prefixIcon: const Icon(Icons.search),
              suffixIcon: _searchController.text.isNotEmpty
                  ? IconButton(
                      icon: const Icon(Icons.clear),
                      onPressed: () {
                        _searchController.clear();
                        _onSearchChanged('');
                      },
                    )
                  : null,
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
              ),
              filled: true,
              fillColor: Colors.grey[100],
            ),
          ),
        ),

        // Loading indicator
        if (_isLoading)
          const Padding(
            padding: EdgeInsets.all(16),
            child: CircularProgressIndicator(),
          ),

        // Error message
        if (_error != null)
          Padding(
            padding: const EdgeInsets.all(16),
            child: Text(
              'Error: $_error',
              style: const TextStyle(color: Colors.red),
            ),
          ),

        // Results count
        if (_searchResults.isNotEmpty)
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Align(
              alignment: Alignment.centerLeft,
              child: Text(
                '${_searchResults.length} results found',
                style: TextStyle(color: Colors.grey[600]),
              ),
            ),
          ),

        // Search results
        Expanded(
          child: _searchResults.isEmpty
              ? Center(
                  child: _searchController.text.isEmpty
                      ? const Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.search, size: 64, color: Colors.grey),
                            SizedBox(height: 16),
                            Text('Type to search for manga'),
                          ],
                        )
                      : _isLoading
                          ? const SizedBox.shrink()
                          : const Text('No results found'),
                )
              : ListView.builder(
                  itemCount: _searchResults.length,
                  itemBuilder: (context, index) {
                    final manga = _searchResults[index];
                    return Card(
                      margin: const EdgeInsets.symmetric(
                        horizontal: 16,
                        vertical: 4,
                      ),
                      child: ListTile(
                        leading: ClipRRect(
                          borderRadius: BorderRadius.circular(4),
                          child: Image.network(
                            manga.thumb,
                            width: 50,
                            height: 70,
                            fit: BoxFit.cover,
                            errorBuilder: (context, error, stackTrace) {
                              return Container(
                                width: 50,
                                height: 70,
                                color: Colors.grey[300],
                                child: const Icon(Icons.image_not_supported),
                              );
                            },
                          ),
                        ),
                        title: Text(
                          manga.name,
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                          style: const TextStyle(fontWeight: FontWeight.bold),
                        ),
                        subtitle: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              manga.author,
                              maxLines: 1,
                              overflow: TextOverflow.ellipsis,
                            ),
                            Text(
                              manga.chapterLatest,
                              style: TextStyle(
                                color: Theme.of(context).colorScheme.primary,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                          ],
                        ),
                        isThreeLine: true,
                        onTap: () => _showDownloadConfirmDialog(manga),
                      ),
                    );
                  },
                ),
        ),
      ],
    );
  }
}
