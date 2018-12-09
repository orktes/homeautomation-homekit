mqtt {
  servers = ["tcp://localhost:1883"]
}
name = "Haaga"
pin = "00102223"

accessory "switch" {
  name = "Vahvistin"
  service "switch" {
    characteristic "on" {
      get = "get('haaga/dra/power')"
      set = "set('haaga/dra/power', value)"
    }
  }
}

accessory "lightbulb" {
  name = "Valot"
  service "lightbulb" { 
    characteristic "on" {
      get = "get('haaga/deconz/groups/3/any_on')"
      set = "set('haaga/deconz/groups/3/on', value)"
    }
    characteristic "brightness" {
      get = "toRange(get('haaga/deconz/groups/3/bri'), [0, 225], [0, 100])"
      set = "set('haaga/deconz/groups/3/bri', toRange(value, [0, 100], [0, 255]))"
    }
  }
}


accessory "switch" {
  name = "Televisio"
  service "switch" {
    characteristic "on" {
      get = "get('haaga/tv/1/power')"
      set = "set('haaga/tv/1/power', value)"
    }
  }
}