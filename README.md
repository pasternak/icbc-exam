# icbc-exam
icbc.com "Book a Road Test" appointment lookup app

A very dumb and simple app to lookup free appointment slots for the icbc road test exam (BC, Canada).

## running

highly recommend running it as a cronjob, every minute to ensure the free spot is found. I have noticed the spots can dissapear in a matter of seconds of appearing in the portal.

### example cronjob:

```
* * * * * /home/pi/icbc-book-drive-test -last-name <last name> -license-number <license number> -keyword <keyword> -start-date 2021-06-18 -location-id 2 -location-id 275 -location-id 276 -pushover-token <pushover.net token> -pushover-user <pushover.net user> -end-date 2021-07-05
```
  
 for more options:
 ```./icbc-book-drive-test -h```
