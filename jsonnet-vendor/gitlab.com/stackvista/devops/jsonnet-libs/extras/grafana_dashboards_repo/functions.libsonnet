{
  local Function = self,

  general: {
    // "Standarize" index key string as "%06d"
    // _idx(i):: ('%06d' % [i]),
    _idx(i):: ('%s' % [i]),

    // Convert array [elem1, elem2, ...] to object as { 000000: elem0, 000001: elem1, ...}
    _to_obj(a):: std.foldl(
      function(x, y) x + y,
      std.mapWithIndex(function(i, v) { [Function.general._idx(i)]: v }, a),
      {}
    ),
  },

  grafana: {
    grid_positioning(panels, default_height=8, default_width=12, total_width=24)::
      local panelsObjects = Function.general._to_obj(panels);
      local numPanels = std.length(panels);
      local panelsPerLine = std.floor(total_width / default_width);

      [
        local numIndex = std.parseInt(index);
        panelsObjects[index] {
          gridPos: {
            h: default_height,
            w: default_width,
            x: std.mod(numIndex, panelsPerLine) * default_width,
            y: std.floor(numIndex / panelsPerLine) * default_height,
          },
        }
        for index in std.objectFields(panelsObjects)
      ],
  },
}
