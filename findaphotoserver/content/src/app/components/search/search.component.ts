import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, NavigationExtras, Router } from '@angular/router';
import { Location } from '@angular/common';
import { Title } from '@angular/platform-browser';


import { BaseSearchComponent } from './base-search.component';
import { SearchRequestBuilder } from '../../models/search.request.builder';

import { DataDisplayer } from '../../providers/data-displayer';
import { FieldsProvider } from '../../providers/fields.provider';
import { NavigationProvider } from '../../providers/navigation.provider';
import { SearchResultsProvider } from '../../providers/search-results.provider';

@Component({
    selector: 'app-search',
    templateUrl: './search.component.html',
    styleUrls: ['./search.component.css']
})

export class SearchComponent extends BaseSearchComponent implements OnInit {
    resultsSearchText: string;


    constructor(
            router: Router,
            route: ActivatedRoute,
            location: Location,
            searchRequestBuilder: SearchRequestBuilder,
            private searchResultsProvider: SearchResultsProvider,
            private navigationProvider: NavigationProvider,
            private displayer: DataDisplayer,
            private fieldsProvider: FieldsProvider,
            private titleService: Title) {
        super('/search', router, route, location, searchRequestBuilder, searchResultsProvider, navigationProvider, fieldsProvider);
    }

    ngOnInit() {
        this.titleService.setTitle('Search - FindAPhoto');
        this.uiState.showSearch = true;
        this.uiState.showResultCount = true;
        this.fieldsProvider.initialize();
        this.navigationProvider.initialize();
        this.searchResultsProvider.initializeRequest(SearchResultsProvider.QueryProperties, 's');

        this.drilldownFromUrl = this.searchResultsProvider.searchRequest.drilldown;

        this._route.queryParams.subscribe(params => {
            if ('q' in params || 't' in params) {
                this.internalSearch(false);
            }
        });
    }

    userSearch() {
        // If the search is new or different, navigate so we can use browser back to get to previous search results
        if (this.resultsSearchText && this.resultsSearchText !== this._searchResultsProvider.searchRequest.searchText) {
            let navigationExtras: NavigationExtras = {
                queryParams: { q: this._searchResultsProvider.searchRequest.searchText }
            };

            this._router.navigate( ['search'], navigationExtras);
            return;
        }

        this.internalSearch(true);
    }

    processSearchResults() {
        this.resultsSearchText = this._searchResultsProvider.searchRequest.searchText;
        if (this.resultsSearchText.length > 0) {
            this.titleService.setTitle(this.resultsSearchText + ' - FindAPhoto');
        } else {
            this.titleService.setTitle(this.resultsSearchText + 'Search all - FindAPhoto');
        }
    }
}
