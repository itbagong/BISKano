import {StyleSheet, Text, View} from 'react-native';
import React from 'react';
import {Card, CheckBox} from '@ui-kitten/components/ui';
import {Helper, Mixins, Typography, Colors} from 'utils';
import moment from 'moment';

type Props = {
  loading: any;
  item: any;
  onChecked: (nextValue: boolean) => void;
  onSelect: any;
};

const Layout = (props: Props) => {
  const {item, onChecked} = props;
  return (
    <Card style={styles.card} onPress={props.onSelect}>
      <View style={styles.content}>
        <View style={styles.contentLeft}>
          <CheckBox
            style={{}}
            checked={item.checked}
            onChange={nextChecked => {
              onChecked(nextChecked);
            }}
          />
          <View>
            <Text
              style={{
                ...Typography.textMdPlusSemiBold,
                color: Colors.SHADES.dark[700],
              }}>
              {item.Group} [{item.SiteID !== '' ? item.SiteID : '-'}]
            </Text>
            {/* <Text style={{...Typography.textMd}}>
              {moment(item.Date).format('DD MMM YYYY')}
            </Text> */}
          </View>
        </View>
        <View style={styles.contentRight}>
          <Text
            style={{
              ...Typography.textMdPlusSemiBold,
              color: Colors.SHADES.dark[700],
            }}>
            {Helper.currencyFormat(item.Total)}
          </Text>
        </View>
      </View>
    </Card>
  );
};

export default Layout;

const styles = StyleSheet.create({
  card: {
    borderRadius: Mixins.scaleSize(8),
    padding: 0,
  },
  content: {
    flexDirection: 'row',
    justifyContent: 'space-between',
    alignItems: 'center',
  },
  contentLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: Mixins.scaleSize(20),
  },
  contentRight: {},
});
