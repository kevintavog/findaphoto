import { Injectable } from '@angular/core';
import { SearchService } from '../services/search.service';

@Injectable()
export class FieldsProvider {
    initialized: boolean = false;
    fields: [string]

    serverError: string;
    nameWithValues: string;
    values: [string];

    constructor(private searchService: SearchService) {
    }

    initialize() {
        this.nameWithValues = null;
        this.values = null;
        if (this.initialized) { return; }

        this.searchService.indexFields().subscribe(
            fieldResults => {
                this.fields = fieldResults.fields;
                this.initialized = true;
            },
            error => {
                this.serverError = error;
            }
        );
    }

    getValuesForField(fieldName: string) {
        this.values = null;
        this.serverError = null;
        this.nameWithValues = fieldName;

        this.searchService.indexFieldValues(fieldName, null, null).subscribe(
            results => {
                this.values = results.values.values;
            },
            error => {
              this.serverError = error;
            }
        );
    }

    getValuesForFieldWithSearch(fieldName: string, searchText: string, drilldown: string) {
        this.values = null;
        this.serverError = null;
        this.nameWithValues = fieldName;

        this.searchService.indexFieldValues(fieldName, searchText, drilldown).subscribe(
            results => {
                this.values = results.values.values;
            },
            error => {
              this.serverError = error;
            }
        );
    }
}
