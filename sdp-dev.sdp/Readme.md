### Standart

### Standart Struktur File (\*model bisa saja berubah)

- app = berisi main dari services
- cid = berisi pipeline CI
- data = ????
- devops = berisi config - config
- sdplogic = berisi custom API yang tidak ada di model (ex: /sdp/{yourmodule}/{customapi})
- sdpmodel = berisi model UI yang bisa di deliver ke FE (formconfig, gridconfig) dan juga CRUD (insert, update, delete, get, gets). ini bersifat auto generate API dan tidak perlu dimasukkan ke sdplogic

### Rule

- usahakan model UI dan model CRUD di **pisah**
  tata cara di pisah yaitu module CRUD di buat 1 model saja. sedangkan model grid 1 model dan model form 1 model.
- usahakan testing
- penamaan usahakan ada min dan huruf kecil semua

### Constributing Guidelines

### SUIM tags

reference : https://github.com/ariefdarmawan/suim

#### Form Setting

- **key** (tag only)
- **obj_title** (string)
- **form_hide_buttons** (bool in number)
- **form_hide_edit_button** (bool in number)
- **form_hide_submit_button** (bool in number)
- **form_hide_cancel_button** (bool in number)
- **form_initial_mode** (string)
- **form_submit_text** (string)
- **form_auto_col** (number)
- **form_section_direction** (string)
- **form_section_size** (number)

#### Grid Setting

- **key** (tag only)
- **grid_keyword** (bool in number) = "add search keyword in field"
- **grid_sortable** (bool in number) = "add sort in field"

#### Main Object

- **obj_go_validator**

#### The Field Form

- **form_allow_add** (bool in number)
- **form_decimal** (bool in number)
- **form_date_format** (string)
- **form** (show | hide)
- **form_pos** (number (Y, X) ex: 1,2)
- **form_section** (string)
- **form_section_width** (string in number)
- **form_section_show_title** (bool in number)
- **form_section_auto_col** (bool in number)
- **form_unit**
- **form_kind** ("radio" | ("bool" | "checkbox") = checkbox | "html" | "password" | "number" ) ) reference type https://www.w3schools.com/tags/tag_input.asp
- **form_disable** (tag only)
- **form_fix_detail** (tag only)
- **form_fix_title** (tag only)
- **form_hint** (string) = for hint
- **form_items** (array and splitter "|" ex: a|b|c) = for item
- **obj_label_field**
- **form_label**
- **form_multiple**
- **form_use_list**
- **form_lookup** (string) = for find value in another service
- **form_placeholder** (string) = for place holder
- **form_length** (MinLength, MaxLength (number)) = for length
- **form_multi_row** (number) = for multi row
- **form_required** (tag only) = for required
- **form_read_only** (number ( 0 / 1 )) = for read only
- **form_read_only_edit** (number ( 0 / 1 )) = for read only on edit
- **form_read_only_new** (number ( 0 / 1 )) = read only on New
- **form_hide_detail** (tag only) = for show detail
- **form_hide_hint** (tag only) = for hidden hint
- **form_hide_title** (tag only) = for show title
- **form_width** (string) = for width form
- **form_space_before** (number) = for space before
- **form_space_after** (number) = for space after
