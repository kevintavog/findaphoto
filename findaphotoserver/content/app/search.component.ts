import { Component, OnInit } from 'angular2/core';
import { Router, ROUTER_DIRECTIVES, RouteParams, Location } from 'angular2/router';

import { SearchRequest } from './search-request';
import { SearchResults,SearchGroup,SearchItem } from './search-results';
import { SearchService } from './search.service';

import { DateStringToLocaleDatePipe } from './datestring-to-localedate.pipe';

@Component({
  selector: 'search',
  templateUrl: 'app/search.component.html',
  styleUrls:  ['app/search.component.css'],
  directives: [ROUTER_DIRECTIVES],
  pipes: [DateStringToLocaleDatePipe]
})

export class SearchComponent implements OnInit {
    private static QueryProperties: string = "id,city,keywords,imageName,createdDate,thumbUrl,slideUrl"
    searchRequest: SearchRequest;
    searchResults: SearchResults;
    resultsSearchText: string;

  constructor(
    private _router: Router,
    private _searchService: SearchService,
    private _routeParams: RouteParams,
    private _location: Location) { }

  ngOnInit() {
    let autoSearch = ("q" in this._routeParams.params)
    let firstItem = 1
    let pageCount = 20

    let searchText = this._routeParams.get("q")
    if (!searchText) {
        searchText = ""
    }

    this.searchRequest = { searchText: searchText, first: firstItem, pageCount: pageCount, properties: SearchComponent.QueryProperties };
    if (autoSearch) {
        this.internalSearch(false)
    }
  }

  userSearch() {
      // If the search is new or different, navigate so we can use browser back to get to previous search results
      if (this.resultsSearchText && this.resultsSearchText != this.searchRequest.searchText) {
          this._router.navigate( ['Search', { q: this.searchRequest.searchText }] );
          return
      }

      this.internalSearch(true)
  }

  internalSearch(updateUrl: boolean) {
      if (updateUrl) {
          this._location.go("/search", "q=" + this.searchRequest.searchText)
      }

      this.searchResults = undefined;
      this._searchService.search(this.searchRequest).subscribe(
          results => {
              this.searchResults = results
              this.resultsSearchText = this.searchRequest.searchText

              let resultIndex = 0
              for (var group of this.searchResults.groups) {
                  group.resultIndex = resultIndex
                  resultIndex += group.items.length
              }
          },
          error => console.log("Handle error: " + error) );
  }
}
