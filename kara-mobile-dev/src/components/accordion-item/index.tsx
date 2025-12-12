import React, {useState} from 'react';
import type {PropsWithChildren} from 'react';
import {StyleSheet, Text, TouchableOpacity, View} from 'react-native';
import {Divider, Icon} from '@ui-kitten/components/ui';
import {Colors, Mixins, Typography} from 'utils';

type AccordionItemPros = PropsWithChildren<{
  title: string;
  header?: any;
}>;

const AccordionItem = ({children, title, header}: AccordionItemPros) => {
  const [expanded, setExpanded] = useState(false);

  function toggleItem() {
    setExpanded(!expanded);
  }

  const body = (
    <>
      <Divider />
      <View style={styles.accordBody}>{children}</View>
    </>
  );
  return (
    <View style={styles.accordContainer}>
      <TouchableOpacity style={styles.accordHeader} onPress={toggleItem}>
        {header ? header : <Text style={styles.accordTitle}>{title}</Text>}
        <Icon
          name={expanded ? 'chevron-up-outline' : 'chevron-down-outline'}
          style={{
            width: Mixins.scaleSize(25),
            height: Mixins.scaleSize(25),
          }}
          color={Colors.SHADES.gray[900]}
        />
      </TouchableOpacity>
      {expanded && body}
    </View>
  );
};
export default AccordionItem;
const styles = StyleSheet.create({
  accordContainer: {
    marginBottom: Mixins.scaleSize(10),
    borderRadius: Mixins.scaleSize(8),
    backgroundColor: '#ffff',
  },
  accordHeader: {
    padding: Mixins.scaleSize(10),
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  accordTitle: {
    ...Typography.textLgSemiBold,
    color: Colors.BLACK,
  },
  accordBody: {
    padding: Mixins.scaleSize(10),
  },
});
