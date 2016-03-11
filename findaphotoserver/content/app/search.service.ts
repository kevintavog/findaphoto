import { SearchRequest } from './search-request';
import { Injectable } from 'angular2/core';
import { Http, RequestOptionsArgs, Response, Headers } from 'angular2/http'
import { Observable } from 'rxjs/Observable';
import 'rxjs/Rx';

@Injectable()
export class SearchService {
 constructor(private _http: Http) {}

  search(request: SearchRequest) {
      var url = "/api/search?q=" + request.searchText + "&properties=" + request.properties + "&first=" + request.first + "&count=" + request.pageCount

      return this._http.get(url)
                  .map(response => response.json())
                  .catch(this.handleError);
  }

  private handleError(response: Response) {
      console.error("Server returned " + response)
      console.error("Server error as json " + response.json())
      let error = response.json()
      if (error != undefined) {
          return Observable.throw(error.errorCode + "; " + error.errorMessage);
      }
      return Observable.throw(response.text());
  }
}
