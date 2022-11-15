# engine

Filter .ics-files using simple YAML for RALF.

## Example

```yaml
---
# name of the filter profile; shown in the RALF-dashboard
name: >
  Rename 'TINF[...][...][...]' to [...]-[...]-[...], e. g. TINF14B1 --> B-14-1 
  and only include courses on Mondays and Tuesdays after 9 AM"

# time to cache a response from an .ics source to prevent rate limiting
cache-duration: 5m

# flows are executed in order
flows:
  # filter out all courses by default.
  # we can filter them in later using the `filters/filter-in` action.
  - do: filters/filter-out
    
  # only include mondays and tuesdays after 10:00
  - if: '(Date.isMonday() or Date.isTuesday()) and Date.isAfter("9:00")'
    then:
      # filter in course
      - do: filters/filter-in
  
      # rename course
      - do: actions/regex-replace
        with:
          match: 'TINF(\d+)([A-Z]+)(\d+)'
          in: [ "DESCRIPTION"]
          replace: '$2-$1-$3'
          
      # btw. you can also nest if-s
      - if: ...
        then:
          - do: ...
        else:
          - ...
...
```

## WIP: Context based actions

This action should put the room into the description of an event

> **Note**: This is a work in progress and does not work at the moment. The syntax can vary in the future.

```yaml
---
name: >
  Write the room into the description
cache-duration: 5m
flows:
  
  # extract Room name from the LOCATION attribute
  - do: actions/get-attribute
    with:
      attribute: "LOCATION"
      into: Room
      
  # check if event had a location specified
  - if: 'Context.Current.Room != ""'
    then:
      
      # prepend Room name to DESCRIPTION attribute
      - do: actions/set-attribute
        with:
          attribute: "DESCRIPTION"
          to: "'[' + Context.Current.Room + '] ' + Event.Description"
...
```