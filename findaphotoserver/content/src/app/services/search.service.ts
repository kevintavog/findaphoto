import { Injectable } from '@angular/core';
import { Http, Response, Headers, ResponseType } from '@angular/http';
import { Observable } from 'rxjs/Observable';
import 'rxjs/Rx';

@Injectable()
export class SearchService {

    constructor(private _http: Http) { }


    searchByText(searchText: string, properties: string, first: number, pageCount: number, drilldown: string) {
        let url = '/api/search?q=' + searchText + '&first=' + first + '&count=' + pageCount + '&properties='
            + properties + '&categories=keywords,tags,placename,date';
        if (drilldown !== undefined && drilldown.length > 0) {
           url += '&drilldown=' + drilldown;
        }

        return this._http.get(url)
            .map(response => response.json())
            .catch(this.handleError);
    }

    searchByDay(month: number, day: number, properties: string, first: number, pageCount: number, random: boolean, drilldown: string) {
        let url = '/api/by-day?month=' + month + '&day=' + day + '&first=' + first + '&count=' + pageCount + '&properties='
            + properties + '&categories=keywords,tags,placename,year';
        if (drilldown !== undefined && drilldown.length > 0) {
            url += '&drilldown=' + drilldown;
        }
        if (random) {
            url += '&random=' + random;
        }

        return this._http.get(url)
            .map(response => response.json())
            .catch(this.handleError);
    }

    searchByLocation(lat: number, lon: number,
            maxKilometers: number, properties: string, 
            first: number, pageCount: number, drilldown: string) {
        let url = '/api/nearby?lat=' + lat + '&lon=' + lon + '&first=' + first + '&count=' + pageCount + '&properties='
            + properties + '&categories=keywords,tags,date';

        if (drilldown && drilldown.length > 0) {
            url += '&drilldown=' + drilldown;
        }

        if (maxKilometers > 0) {
            url += '&maxKilometers=' + maxKilometers;
        }

        let headers = new Headers();
        headers.append('Content-Type', 'application/json');
        headers.append('Accept', 'application/json');

        return this._http.get(url, { headers: headers })
            .map(response => response.json())
            .catch(this.handleError);
    }

    indexStats(properties: string) {
        let url = '/api/index?properties=' + properties;
        return this._http.get(url)
            .map(response => response.json())
            .catch(this.handleError);
    }

    indexFields() {
        let url = '/api/index?properties=fields';
        return this._http.get(url)
            .map(response => response.json())
            .catch(this.handleError);
    }

    indexFieldValues(fieldName: [string], query: string, drilldown: string, maxCount: number) {
        let url = '/api/index/fieldvalues?fields=' + fieldName.join(',');

        if (maxCount != null) {
            url += '&max=' + maxCount;
        }

        if (query != null && query.length > 0) {
            url += '&q=' + query;
        }
        if (drilldown != null && drilldown.length > 0) {
            url += '&drilldown=' + drilldown;
        }

        return this._http.get(url)
            .map(response => response.json())
            .catch(this.handleError);
    }

    indexFieldValuesByDay(fieldName: [string], month: number, day: number, drilldown: string, maxCount: number) {
        let url = '/api/index/fieldvalues?fields=' + fieldName.join(',') + '&month=' + month + '&day=' + day;

        if (maxCount != null) {
            url += '&max=' + maxCount;
        }

        if (drilldown != null && drilldown.length > 0) {
            url += '&drilldown=' + drilldown;
        }

        return this._http.get(url)
            .map(response => response.json())
            .catch(this.handleError);
    }

    mediaSource(id: string) {
        let url = '/api/media/' + id;
        return this._http.get(url)
            .map(response => response.json())
            .catch(this.handleError);
    }

    private handleError(response: Response) {
        if (response.type === ResponseType.Error) {
            return Observable.throw('Server not accessible');
        }

        if (response.status >= 500) {
            return Observable.throw('Server error: ' + response.status + '; ' + response.text());
        }

        let error = response.json();
        if (!error) {
            return Observable.throw('The server returned: ' + response.text());
        }

        return Observable.throw('The server failed with: ' + error.errorCode + '; ' + error.errorMessage);
    }
}
