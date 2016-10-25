import { Injectable } from '@angular/core';
import { Http, RequestOptionsArgs, Response, Headers, ResponseType } from '@angular/http'
import { Observable } from 'rxjs/Observable';
import 'rxjs/Rx';

import { SearchRequest } from '../models/search-request';

@Injectable()
export class SearchService {

    constructor(private _http: Http) { }


    searchByText(searchText: string, properties: string, first: number, pageCount: number, drilldown: string) {
       var url = "/api/search?q=" + searchText + "&first=" + first + "&count=" + pageCount + "&properties=" + properties + "&categories=keywords,placename,date"
       if (drilldown != undefined && drilldown.length > 0) {
           url += "&drilldown=" + drilldown
       }

       return this._http.get(url)
                   .map(response => response.json())
                   .catch(this.handleError);
    }

    searchByDay(month: number, day: number, properties: string, first: number, pageCount: number, random: boolean, drilldown: string) {
       var url = "/api/by-day?month=" + month + "&day=" + day + "&first=" + first + "&count=" + pageCount + "&properties=" + properties + "&categories=keywords,placename,year"
       if (drilldown != undefined && drilldown.length > 0) {
           url += "&drilldown=" + drilldown
       }
       if (random) {
           url += "&random=" + random
       }

       return this._http.get(url)
                   .map(response => response.json())
                   .catch(this.handleError);
    }

    searchByLocation(lat: number, lon: number, properties: string, first: number, pageCount: number, drilldown: string) {
       var url = "/api/nearby?lat=" + lat + "&lon=" + lon + "&first=" + first + "&count=" + pageCount + "&properties=" + properties + "&categories=keywords,date"
       if (drilldown != undefined && drilldown.length > 0) {
           url += "&drilldown=" + drilldown
       }

       let headers = new Headers();
       headers.append('Content-Type', 'application/json');
       headers.append('Accept', 'application/json');

       return this._http.get(url, { headers: headers })
                     .map(response => response.json())
                     .catch(this.handleError);
    }

    private handleError(response: Response) {
       if (response.type == ResponseType.Error) {
           return Observable.throw("Server not accessible");
       }

       let error = response.json()
       if (!error) {
           return Observable.throw("The server returned: " + response.text());
       }

       return Observable.throw("The server failed with: " + error.errorCode + "; " + error.errorMessage);
    }

}