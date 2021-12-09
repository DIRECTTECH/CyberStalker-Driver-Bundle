CREATE OR REPLACE PROCEDURE pix.init_object_by_schema(schema_id uuid, object_id uuid)
 LANGUAGE plpgsql
AS $procedure$
DECLARE
		v_prop pix.schema_properties;
		v_control pix.schema_controls;
		v_object_id ALIAS FOR $2;
		v_schema_id_param ALIAS FOR $1;
		v_schema_id uuid;
		v_linked_object_id uuid;
		v_schema_to_create_id uuid;
		v_object_to_delete_id uuid;
		v_editorgroup uuid;
		v_usergroup uuid;
		v_readergroup uuid;
		v_schema_name pix.schemas.name%TYPE;
		v_object_record pix.objects;
BEGIN

      IF v_schema_id_param IS NULL THEN
         SELECT objects.schema_id INTO v_schema_id FROM pix.objects WHERE objects.id = v_object_id;
         IF NOT FOUND THEN
              RETURN;
         END IF;
	  ELSE
	     v_schema_id = v_schema_id_param;
	  END IF;

	  SELECT editorgroup, usergroup, readergroup FROM pix.objects WHERE objects.id = v_object_id
	  	INTO v_editorgroup, v_usergroup, v_readergroup;

	  SELECT * FROM pix.objects WHERE objects.id = v_object_id
	  	INTO v_object_record;

	  RAISE NOTICE 'object name: %', v_object_record.name;

--
-- sync Properties
--
      FOR v_prop IN SELECT * FROM pix.schema_properties WHERE schema_properties.schema_id = v_schema_id

         LOOP
--
--          UPSERT properties from SCHEMA
--
				INSERT INTO pix.object_properties (object_id,group_name,property,type_id,value,index,stealth,editorgroup,usergroup,readergroup)
					VALUES (v_object_id,
						v_prop.group_name,
					    v_prop.property,
						v_prop.type_id,
					    v_prop.default_value,
						v_prop.index,
						v_prop.stealth,
						v_editorgroup,
						v_usergroup,
						v_readergroup)
					ON CONFLICT ON CONSTRAINT object_groupname_property_constraint
					DO
					UPDATE SET
--						group_name = v_prop.group_name,
--						property = v_prop.property,
						type_id = v_prop.type_id,
                        --- value = v_prop.default_value,
                        -- value = (CASE WHEN (v_prop.default_value IS NOT NULL) THEN v_prop.default_value END),
                        -- value = (CASE WHEN (v_prop.default_value IS NULL) THEN value ELSE v_prop.default_value END),
						stealth = v_prop.stealth;
--						tags = (select array_agg(distinct tag) from (select unnest(array[object_properties.tags]) as tag union
--								select unnest(array[v_prop.m_tags]) as tag) t);
--						value = v_prop.default_value,
--						flag = v_prop.default_flag;

                -- UPDATE pix.object_properties SET value = v_prop.default_value WHERE v_prop.default_value IS NOT NULL AND values IS NOT NULL;
		 END LOOP;
-- UPDATE propertie values if not set and
      -- FOR v_prop IN SELECT * FROM pix.schema_properties WHERE schema_properties.schema_id = v_schema_id
         -- LOOP
                -- UPDATE pix.object_properties AS op SET value = v_prop.default_value
                -- WHERE v_prop.default_value IS NOT NULL
                    -- AND op.value IS NULL;
		 -- END LOOP;

--
--          	DELETE removed properties from OBJECT_PROPERTIES
--
				DELETE FROM pix.object_properties
			      WHERE object_properties.object_id = v_object_id
			      AND (object_properties.group_name,object_properties.property,object_properties.index)
					NOT IN (SELECT group_name,property,index FROM pix.schema_properties
								 WHERE schema_properties.schema_id = v_schema_id);

--
-- Schema2Schema support logics.
--

--
--	  Select all connected schemas which don't have its object presented in objects2objects relation and create such objects.
--    Then connect them.
--
	  FOR v_schema_to_create_id IN SELECT schemas.id FROM pix.schemas
	  		WHERE schemas.parent_schema_id = v_schema_id AND schemas.id NOT IN
			(SELECT objects.schema_id FROM pix.objects_to_objects, pix.objects
	  		WHERE objects_to_objects.object1_id = v_object_id AND objects_to_objects.object2_id = objects.id
				AND objects_to_objects.forced > 0)
	  LOOP
	  		SELECT schemas.name FROM pix.schemas WHERE schemas.id = v_schema_to_create_id INTO v_schema_name;
	  		INSERT INTO pix.objects
					(name, enabled, schema_id, editorgroup, usergroup, readergroup)
					VALUES (v_schema_name,v_object_record.enabled,v_schema_to_create_id, v_object_record.editorgroup,
						   		v_object_record.usergroup, v_object_record.readergroup)
					RETURNING id INTO v_linked_object_id;
			UPDATE pix.objects SET editorgroup = '11111111-1111-1111-1111-111111111111' WHERE id = v_linked_object_id;
			INSERT INTO pix.objects_to_objects (object1_id, object2_id, forced) VALUES (v_object_id, v_linked_object_id, 1);
	  END LOOP;

--
--	  Select all connected objects which don't have matching schema presented in schemas2schemas relation and delete such objects.
--

	  FOR v_object_to_delete_id IN SELECT object2_id FROM pix.objects_to_objects, pix.objects
	  		WHERE object1_id = v_object_id AND object2_id = objects.id AND objects_to_objects.forced > 0 AND objects.schema_id NOT IN
			(SELECT schemas.id FROM pix.schemas
	  		WHERE schemas.parent_schema_id = v_schema_id)
	  LOOP
	  		DELETE FROM pix.objects WHERE objects.id = v_object_to_delete_id;
	  END LOOP;

END
$procedure$
;
