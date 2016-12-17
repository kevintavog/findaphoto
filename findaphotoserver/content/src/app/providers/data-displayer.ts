import { Injectable } from '@angular/core';
import { SearchItem } from '../models/search-results';

interface DegreesMinutesSeconds {
    degrees: number;
    minutes: number;
    seconds: number;
}

@Injectable()
export class DataDisplayer {

    itemYear(item: SearchItem) {
        let date = this.getItemDate(item);
        if (date != null) {
            return date.getFullYear();
        }
        return -1;
    }

    itemMonth(item: SearchItem) {
        let date = this.getItemDate(item);
        if (date != null) {
            return date.getMonth() + 1;
        }
        return -1;
    }

    itemDay(item: SearchItem) {
        let date = this.getItemDate(item);
        if (date != null) {
            return date.getDate();
        }
        return -1;
    }

    getItemDate(item: SearchItem) {
        if (item.createdDate != null) {
            let date = item.createdDate;
            if (typeof item.createdDate === 'string') {
                date = new Date(item.createdDate);
            }
            return date;
        }
        return undefined;
    }

    getItemLocaleDate(item: SearchItem) {
        if (item.createdDate != null) {
            return new Date(item.createdDate).toLocaleDateString();
        }
        return undefined;
    }

    getItemLocaleDateAndTime(item: SearchItem) {
        if (item.createdDate != null) {
            let d = new Date(item.createdDate);
            return d.toLocaleDateString() + '  ' + d.toLocaleTimeString();
        }
        return undefined;
    }

    dateToLocaleDateAndTime(date: Date) {
        if (date != null) {
            let d = new Date(date);
            return d.toLocaleDateString() + '  ' + d.toLocaleTimeString();
        }
        return undefined;
    }

    lonDms(item: SearchItem) {
        return this.longitudeDms(item.longitude);
    }

    longitudeDms(longitude: number) {
        return this.convertToDms(longitude, ['E', 'W']);
    }

    latDms(item: SearchItem) {
        return this.latitudeDms(item.latitude);
    }

    latitudeDms(latitude: number) {
        return this.convertToDms(latitude, ['N', 'S']);
    }


    convertToDms(degrees: number, refValues: string[]): string {
        let dms = this.degreesToDms(degrees);
        let ref = refValues[0];
        if (dms.degrees < 0) {
            ref = refValues[1];
            dms.degrees *= -1;
        }
        return dms.degrees + '° ' + dms.minutes + '\' ' + dms.seconds.toFixed(2) + '\" ' + ref;
    }

    degreesToDms(degrees: number): DegreesMinutesSeconds {

        let d = degrees;
        if (d < 0) {
            d = Math.ceil(d);
        } else {
            d = Math.floor(d);
        }

        let minutesSeconds = Math.abs(degrees - d) * 60.0;
        let m = Math.floor(minutesSeconds);
        let s = (minutesSeconds - m) * 60.0;

        return {
            degrees: d,
            minutes: m,
            seconds: s};
    }

}
