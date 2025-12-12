/* eslint-disable @typescript-eslint/no-shadow */
/* eslint-disable react/no-unstable-nested-components */
/* eslint-disable react-hooks/exhaustive-deps */
/* eslint-disable react-native/no-inline-styles */
import {FlatList, RefreshControl, SafeAreaView, StyleSheet} from 'react-native';
import React from 'react';
import {Colors, Mixins} from 'utils';
import {useActions, useAppState} from '@overmind/index';
import {default as theme} from '../../custom-theme.json';
import CardAttendanceHistory from '@components/module/card-history-attendance';

type Props = {
  navigation: any;
  route: any;
};

const Layout = (props: Props) => {
  const {} = props;
  const {} = useAppState();
  const {getListHistory} = useActions();
  const [data, setData] = React.useState();
  const init = () => {
    setRefreshing(true);
    getListHistory({sort: ['-_id']}).then(r => {
      setData(r);
      setRefreshing(false);
    });
  };

  React.useEffect(() => {
    init();
    return () => {};
  }, []);

  const [refreshing, setRefreshing] = React.useState(false);

  const onRefresh = React.useCallback(() => {
    init();
  }, []);

  const RenderItem = ({item}: any) => {
    return <CardAttendanceHistory data={item} />;
  };
  return (
    <>
      <SafeAreaView style={styles.container}>
        <FlatList
          data={data}
          renderItem={listrender => {
            return <RenderItem item={listrender.item} />;
          }}
          refreshControl={
            <RefreshControl refreshing={refreshing} onRefresh={onRefresh} />
          }
          keyExtractor={(item, index) => index.toString()}
        />
      </SafeAreaView>
    </>
  );
};

export default Layout;

const styles = StyleSheet.create({
  container: {
    flex: 1,
    paddingTop: Mixins.scaleSize(20),
    paddingHorizontal: Mixins.scaleSize(14),
  },
  card: {
    marginBottom: Mixins.scaleSize(20),
    marginHorizontal: Mixins.scaleSize(20),
    borderRadius: 10,
    borderColor: theme['color-primary-200'],
    shadowColor: Colors.SHADES.dark[100],
    shadowOffset: {width: 0, height: 0},
    shadowOpacity: 70,
  },
});
